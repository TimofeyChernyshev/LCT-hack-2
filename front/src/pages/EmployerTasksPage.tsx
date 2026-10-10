import { useEffect, useState, type FormEvent } from "react";
import { api } from "../api/client";
import type { Named, TaskAnswer } from "../api/types";
import { Button, Empty, Field, Notice, PageTitle, Select, TextArea, TextInput } from "../components/ui";
import { categoryFor, loadCatalog } from "../lib/catalog";
import { formatWhen } from "../lib/format";

export function EmployerTasksPage() {
  const [specs, setSpecs] = useState<Named[]>([]);
  const [grades, setGrades] = useState<Named[]>([]);
  const [categories, setCategories] = useState<Awaited<ReturnType<typeof loadCatalog>>["categories"]>([]);
  const [answers, setAnswers] = useState<TaskAnswer[]>([]);
  const [form, setForm] = useState({ title: "", body: "", specializationId: "", gradeId: "" });
  const [inviteKey, setInviteKey] = useState("");
  const [offer, setOffer] = useState({ salaryMin: "", salaryMax: "", message: "", contact: "" });
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [pending, setPending] = useState(false);

  async function loadAnswers() {
    setAnswers(await api<TaskAnswer[]>("testing", "/periodic-tasks"));
  }

  useEffect(() => {
    loadCatalog()
      .then((catalog) => {
        setSpecs(catalog.specializations);
        setGrades(catalog.grades);
        setCategories(catalog.categories);
      })
      .catch(() => undefined);
    loadAnswers().catch((reason: unknown) => {
      setError(reason instanceof Error ? reason.message : "Ответы не загрузились");
    });
  }, []);

  async function create(event: FormEvent) {
    event.preventDefault();
    const category = categoryFor({ specializations: specs, grades, technologies: [], categories }, form.specializationId, form.gradeId);
    if (!category) {
      setError("Для этой пары специализации и грейда нет категории.");
      return;
    }
    setPending(true);
    setError("");
    setInfo("");
    try {
      await api("testing", "/periodic-tasks", {
        method: "POST",
        body: JSON.stringify({ categoryId: category.id, title: form.title.trim(), body: form.body.trim() }),
      });
      setForm({ title: "", body: "", specializationId: "", gradeId: "" });
      setInfo("Задание опубликовано.");
      await loadAnswers();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Задание не создано");
    } finally {
      setPending(false);
    }
  }

  async function react(item: TaskAnswer, reaction: "confirmed" | "thanks" | "invited") {
    setError("");
    try {
      await api("testing", "/periodic-tasks/reactions", {
        method: "POST",
        body: JSON.stringify({ taskId: item.taskId, userId: item.userId, reaction }),
      });
      await loadAnswers();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Реакция не сохранилась");
    }
  }

  async function invite(event: FormEvent, item: TaskAnswer) {
    event.preventDefault();
    const salaryMin = Number(offer.salaryMin);
    const salaryMax = Number(offer.salaryMax);
    if (!Number.isFinite(salaryMin) || !Number.isFinite(salaryMax) || salaryMin > salaryMax) {
      setError("Вилка обязательна: «от» не больше «до».");
      return;
    }
    if (!offer.message.trim() || !offer.contact.trim()) {
      setError("Нужны текст приглашения и способ связи.");
      return;
    }
    setPending(true);
    setError("");
    try {
      await api("interaction", "/me/invitations", {
        method: "POST",
        body: JSON.stringify({
          candidateUserId: item.userId,
          message: offer.message.trim(),
          salaryMin,
          salaryMax,
          currency: "RUB",
          contactChannel: "email",
          employerContact: offer.contact.trim(),
        }),
      });
      await react(item, "invited");
      setInviteKey("");
      setInfo("Приглашение отправлено.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Приглашение не отправилось");
    } finally {
      setPending(false);
    }
  }

  const reactionText = {
    confirmed: "Результат подтверждён",
    thanks: "Поблагодарили",
    invited: "Приглашение отправлено",
  };

  return (
    <div className="grid gap-8">
      <PageTitle title="Мини-задачи" />
      {error ? <Notice>{error}</Notice> : null}
      {info ? <Notice tone="ok">{info}</Notice> : null}
      <form className="grid max-w-xl gap-4" onSubmit={create}>
        <Field label="Название">
          <TextInput required value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} />
        </Field>
        <Field label="Условие">
          <TextArea required value={form.body} onChange={(event) => setForm({ ...form, body: event.target.value })} />
        </Field>
        <div className="grid gap-4 md:grid-cols-2">
          <Field label="Специализация">
            <Select required value={form.specializationId} onChange={(event) => setForm({ ...form, specializationId: event.target.value })}>
              <option value="">Выберите</option>
              {specs.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </Select>
          </Field>
          <Field label="Грейд">
            <Select required value={form.gradeId} onChange={(event) => setForm({ ...form, gradeId: event.target.value })}>
              <option value="">Выберите</option>
              {grades.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </Select>
          </Field>
        </div>
        <p className="text-sm text-dim">Задание увидят соискатели с этой специализацией и этим грейдом.</p>
        <Button type="submit" disabled={pending}>{pending ? "Публикуем" : "Опубликовать"}</Button>
      </form>
      <section>
        <h2 className="font-display text-2xl">Ответы</h2>
        {answers.length === 0 ? <div className="mt-4"><Empty title="Ответов нет" text="Они появятся, когда соискатель отправит решение или подход." /></div> : (
          <ul className="mt-4 divide-y divide-line border-y border-line">
            {answers.map((item) => {
              const key = `${item.taskId}-${item.userId}`;
              return (
              <li key={key} className="grid gap-3 py-4">
                <p className="text-sm text-dim">{item.taskTitle} · {item.kind === "approach" ? "Подход" : "Решение"} · {formatWhen(item.submittedAt)}</p>
                <p>{item.email}</p>
                <p className="whitespace-pre-wrap">{item.answer}</p>
                {item.reaction ? <p className="text-sm">{reactionText[item.reaction]}</p> : null}
                <div className="flex flex-wrap gap-2">
                  <Button type="button" variant="ghost" onClick={() => void react(item, "confirmed")}>Подтвердить</Button>
                  <Button type="button" variant="ghost" onClick={() => void react(item, "thanks")}>Поблагодарить</Button>
                  <Button type="button" variant="ghost" onClick={() => setInviteKey(inviteKey === key ? "" : key)}>Позвать на работу</Button>
                </div>
                {inviteKey === key ? (
                  <form className="grid gap-3 border border-line p-3" onSubmit={(event) => void invite(event, item)}>
                    <div className="grid gap-3 md:grid-cols-2">
                      <Field label="Вилка от, ₽"><TextInput required inputMode="numeric" value={offer.salaryMin} onChange={(event) => setOffer({ ...offer, salaryMin: event.target.value })} /></Field>
                      <Field label="Вилка до, ₽"><TextInput required inputMode="numeric" value={offer.salaryMax} onChange={(event) => setOffer({ ...offer, salaryMax: event.target.value })} /></Field>
                    </div>
                    <Field label="Приглашение"><TextArea required value={offer.message} onChange={(event) => setOffer({ ...offer, message: event.target.value })} /></Field>
                    <Field label="Как связаться"><TextInput required value={offer.contact} onChange={(event) => setOffer({ ...offer, contact: event.target.value })} /></Field>
                    <Button type="submit" disabled={pending}>Отправить приглашение</Button>
                  </form>
                ) : null}
              </li>
              );
            })}
          </ul>
        )}
      </section>
    </div>
  );
}
