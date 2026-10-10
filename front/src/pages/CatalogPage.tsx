import { useEffect, useState, type FormEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import type { CandidateCard, Invitation, Named, SearchHit, SearchPage, Technology } from "../api/types";
import { Button, Empty, Field, Modal, Notice, PageTitle, Select, Skeleton, TextArea, TextInput } from "../components/ui";
import { loadCatalog } from "../lib/catalog";
import { byId } from "../lib/format";

export function CatalogPage() {
  const [params, setParams] = useSearchParams();
  const [specs, setSpecs] = useState<Named[]>([]);
  const [grades, setGrades] = useState<Named[]>([]);
  const [techs, setTechs] = useState<Technology[]>([]);
  const [page, setPage] = useState<SearchPage | null>(null);
  const [selected, setSelected] = useState<SearchHit | null>(null);
  const [card, setCard] = useState<CandidateCard | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [offer, setOffer] = useState(false);
  const [pending, setPending] = useState(false);
  const [draft, setDraft] = useState({ salaryMin: "", salaryMax: "", message: "", contactChannel: "telegram", employerContact: "" });

  const specializationId = params.get("specializationId") ?? "";
  const gradeId = params.get("gradeId") ?? "";
  const hasFsp = params.get("hasFsp") === "true";
  const stack = params.getAll("stack");
  const asked = byId(grades, gradeId);
  const visible = (page?.items ?? []).filter((item) => {
    if (!asked?.rank) return true;
    const rank = byId(grades, item.gradeId)?.rank;
    return rank === asked.rank || rank === asked.rank + 1;
  });

  useEffect(() => {
    loadCatalog()
      .then((catalog) => {
        setSpecs(catalog.specializations);
        setGrades(catalog.grades);
        setTechs(catalog.technologies);
      })
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    const query = new URLSearchParams();
    if (specializationId) query.set("specializationId", specializationId);
    if (hasFsp) query.set("hasFsp", "true");
    stack.forEach((id) => query.append("stack", id));
    query.set("limit", "20");
    let live = true;
    setLoading(true);
    api<SearchPage>("search", `/candidates?${query.toString()}`)
      .then((value) => {
        if (!live) return;
        setPage(value);
        setError("");
      })
      .catch((reason: unknown) => {
        if (live) setError(reason instanceof Error ? reason.message : "Поиск не ответил");
      })
      .finally(() => {
        if (live) setLoading(false);
      });
    return () => {
      live = false;
    };
  }, [specializationId, gradeId, hasFsp, stack.join("|")]);

  useEffect(() => {
    if (!selected) {
      setCard(null);
      return;
    }
    let live = true;
    api<CandidateCard>("candidate", `/candidates/${selected.userId}`)
      .then((value) => {
        if (live) setCard(value);
      })
      .catch(() => {
        if (live) setCard(null);
      });
    return () => {
      live = false;
    };
  }, [selected]);

  function update(next: Record<string, string | string[] | boolean>) {
    const query = new URLSearchParams(params);
    for (const [key, value] of Object.entries(next)) {
      query.delete(key);
      if (Array.isArray(value)) value.forEach((item) => query.append(key, item));
      else if (typeof value === "boolean") {
        if (value) query.set(key, "true");
      } else if (value) query.set(key, value);
    }
    setParams(query);
  }

  async function sendOffer(event: FormEvent) {
    event.preventDefault();
    if (!selected) return;
    const salaryMin = Number(draft.salaryMin);
    const salaryMax = Number(draft.salaryMax);
    if (!Number.isFinite(salaryMin) || !Number.isFinite(salaryMax) || salaryMin > salaryMax || salaryMin < 0) {
      setError("Вилка обязательна: «от» и «до» в рублях, от не больше до.");
      return;
    }
    if (!draft.employerContact.trim() || !draft.message.trim()) {
      setError("Нужны описание роли и канал связи.");
      return;
    }
    setPending(true);
    setError("");
    try {
      await api<Invitation>("interaction", "/me/invitations", {
        method: "POST",
        body: JSON.stringify({
          candidateUserId: selected.userId,
          message: draft.message,
          salaryMin,
          salaryMax,
          currency: "RUB",
          contactChannel: draft.contactChannel,
          employerContact: draft.employerContact.trim(),
        }),
      });
      setOffer(false);
      setDraft({ salaryMin: "", salaryMax: "", message: "", contactChannel: "telegram", employerContact: "" });
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Приглашение не отправилось");
    } finally {
      setPending(false);
    }
  }

  const contactsHidden = !card?.contacts || card.contacts.masked !== false;

  return (
    <div className="grid gap-6">
      <PageTitle title="Каталог" text="Контакты скрыты, пока кандидат не примет приглашение." />
      {error ? <Notice>{error}</Notice> : null}
      <div className="grid gap-3 md:grid-cols-4">
        <Field label="Специализация">
          <Select value={specializationId} onChange={(event) => update({ specializationId: event.target.value })}>
            <option value="">Все</option>
            {specs.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </Select>
        </Field>
        <Field label="Грейд" hint={gradeId ? "Также показываем на один уровень выше." : undefined}>
          <Select value={gradeId} onChange={(event) => update({ gradeId: event.target.value })}>
            <option value="">Все</option>
            {grades.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </Select>
        </Field>
        <label className="mt-7 flex min-h-11 items-center gap-2 text-sm">
          <input type="checkbox" checked={hasFsp} onChange={(event) => update({ hasFsp: event.target.checked })} />
          Есть история ФСП
        </label>
      </div>
      <div className="flex flex-wrap gap-2">
        {techs.slice(0, 18).map((item) => {
          const on = stack.includes(item.id);
          return (
            <button
              key={item.id}
              type="button"
              className={`min-h-11 border px-3 text-sm ${on ? "border-accent bg-deep" : "border-line"}`}
              onClick={() => update({ stack: on ? stack.filter((id) => id !== item.id) : [...stack, item.id] })}
            >
              {item.name}
            </button>
          );
        })}
      </div>
      <div className="grid gap-6 lg:grid-cols-[1fr_1fr]">
        <div>
          {loading ? <Skeleton className="h-48" /> : null}
          {!loading && visible.length === 0 ? (
            <Empty title="Никого не нашли" text="Ослабьте фильтр." />
          ) : null}
          <ul className="divide-y divide-line border-y border-line">
            {visible.map((item) => {
              const above = Boolean(asked?.rank && byId(grades, item.gradeId)?.rank === asked.rank + 1);
              return (
              <li key={item.userId}>
                <button type="button" className={`w-full px-2 py-4 text-left ${selected?.userId === item.userId ? "bg-raised" : ""} ${above ? "bg-deep/40" : ""}`} onClick={() => setSelected(item)}>
                  <span className="flex items-baseline justify-between gap-3">
                    <span>{item.displayName || "Кандидат"}</span>
                    <span className="text-dim">{Math.round(item.score)}</span>
                  </span>
                  <span className={`mt-1 block text-sm ${above ? "text-salary" : "text-dim"}`}>
                    {byId(specs, item.specializationId)?.name ?? "специализация"} · {byId(grades, item.gradeId)?.name ?? "грейд"}
                    {above ? " · на уровень выше" : ""}
                    {typeof item.testScore === "number" ? ` · тест ${Math.round(item.testScore)}` : ""}
                  </span>
                </button>
              </li>
              );
            })}
          </ul>
        </div>
        <aside className="border border-line p-4">
          {!selected ? <p className="text-dim">Выберите человека в списке.</p> : null}
          {selected ? (
            <div className="grid gap-4">
              <h2 className="font-display text-3xl leading-none">{card?.displayName || selected.displayName || "Кандидат"}</h2>
              <p className={asked?.rank && byId(grades, selected.gradeId)?.rank === asked.rank + 1 ? "text-salary" : undefined}>
                {byId(grades, selected.gradeId)?.name ?? "Грейд"} · {byId(specs, selected.specializationId)?.name ?? "направление"}
                {asked?.rank && byId(grades, selected.gradeId)?.rank === asked.rank + 1 ? " · на уровень выше" : ""}
              </p>
              <p className="border border-line bg-deep px-3 py-3 text-sm">
                {selected.explanation || selected.reasons.join(". ") || "Объяснение ещё не пришло."}
              </p>
              <p>
                {(selected.fspAchievementsCount ?? card?.fsp?.achievementsCount ?? 0) > 0
                  ? `ФСП: ${selected.fspAchievementsCount ?? card?.fsp?.achievementsCount} достижений${selected.fspBestPlace ? `, лучшее место ${selected.fspBestPlace}` : ""}`
                  : "Истории ФСП нет"}
              </p>
              <p className="text-dim">
                {contactsHidden
                  ? "Контакты скрыты до подтверждения кандидатом"
                  : [card?.contacts?.email, card?.contacts?.phone, card?.contacts?.telegram].filter(Boolean).join(", ") || "Контакты пустые"}
              </p>
              <Button onClick={() => setOffer(true)}>Пригласить</Button>
            </div>
          ) : null}
        </aside>
      </div>
      {offer && selected ? (
        <Modal title="Приглашение" onClose={() => setOffer(false)}>
          <form className="grid gap-4" onSubmit={sendOffer}>
            <Field label="От, ₽" hint="До вычета НДФЛ, поле обязательно">
              <TextInput required inputMode="numeric" value={draft.salaryMin} onChange={(event) => setDraft({ ...draft, salaryMin: event.target.value })} />
            </Field>
            <Field label="До, ₽">
              <TextInput required inputMode="numeric" value={draft.salaryMax} onChange={(event) => setDraft({ ...draft, salaryMax: event.target.value })} />
            </Field>
            <Field label="Роль и задачи">
              <TextArea required value={draft.message} onChange={(event) => setDraft({ ...draft, message: event.target.value })} />
            </Field>
            <Field label="Канал">
              <Select value={draft.contactChannel} onChange={(event) => setDraft({ ...draft, contactChannel: event.target.value })}>
                <option value="telegram">Telegram</option>
                <option value="email">Почта</option>
                <option value="phone">Телефон</option>
                <option value="platform">На платформе</option>
              </Select>
            </Field>
            <Field label="Как с вами связаться">
              <TextInput required value={draft.employerContact} onChange={(event) => setDraft({ ...draft, employerContact: event.target.value })} />
            </Field>
            <Button type="submit" disabled={pending}>{pending ? "Отправляем" : "Отправить"}</Button>
          </form>
        </Modal>
      ) : null}
    </div>
  );
}
