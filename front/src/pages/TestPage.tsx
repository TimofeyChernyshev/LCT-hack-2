import { useEffect, useMemo, useState } from "react";
import hljs from "highlight.js/lib/core";
import javascript from "highlight.js/lib/languages/javascript";
import python from "highlight.js/lib/languages/python";
import sql from "highlight.js/lib/languages/sql";
import { api } from "../api/client";
import type { CategoryState, FspState, GradeChanges, Named, SessionItem, SessionResult, TestSession } from "../api/types";
import { Button, Empty, Field, Notice, PageTitle, Select, TextArea } from "../components/ui";
import { categoryFor, loadCatalog, type Catalog } from "../lib/catalog";
import { byId, formatWhen } from "../lib/format";

hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("python", python);
hljs.registerLanguage("sql", sql);

type Draft = string | string[];

function paint(body: string, type: string) {
  const language = type === "sql" ? "sql" : type === "code" ? "javascript" : type === "regex" ? "python" : "";
  if (!language) return null;
  return hljs.highlight(body, { language }).value;
}

function left(expiresAt?: string | null) {
  if (!expiresAt) return null;
  const diff = new Date(expiresAt).getTime() - Date.now();
  if (Number.isNaN(diff)) return null;
  const total = Math.max(0, Math.floor(diff / 1000));
  const minutes = Math.floor(total / 60);
  const seconds = total % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function draftOf(item: SessionItem, choice: string[], text: string): Draft {
  const open = item.type !== "single_choice" && item.type !== "multi_choice";
  if (open || !item.options?.length) return text.trim();
  return item.type === "single_choice" ? (choice[0] ?? "") : choice;
}

function filled(draft: Draft) {
  return Array.isArray(draft) ? draft.length > 0 : draft !== "";
}

function sameDraft(a: Draft | undefined, b: Draft) {
  if (a === undefined) return false;
  if (Array.isArray(a) || Array.isArray(b)) {
    const left = Array.isArray(a) ? a : [a];
    const right = Array.isArray(b) ? b : [b];
    return left.length === right.length && left.every((id, index) => id === right[index]);
  }
  return a === b;
}

export function TestPage() {
  const [catalog, setCatalog] = useState<Catalog | null>(null);
  const [tracks, setTracks] = useState<{ specializationId: string; gradeId: string }[]>([]);
  const [specializationId, setSpecializationId] = useState("");
  const [gradeId, setGradeId] = useState("");
  const [session, setSession] = useState<TestSession | null>(null);
  const [index, setIndex] = useState(0);
  const [choice, setChoice] = useState<string[]>([]);
  const [text, setText] = useState("");
  const [answers, setAnswers] = useState<Record<string, Draft>>({});
  const [result, setResult] = useState<SessionResult | null>(null);
  const [fsp, setFsp] = useState<FspState | null>(null);
  const [cooldown, setCooldown] = useState<GradeChanges | null>(null);
  const [clock, setClock] = useState("");
  const [error, setError] = useState("");
  const [confirmFinish, setConfirmFinish] = useState(false);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    loadCatalog().then(setCatalog).catch((reason: unknown) => {
      setError(reason instanceof Error ? reason.message : "Справочники недоступны");
    });
    api<CategoryState>("candidate", "/me/category")
      .then((state) => {
        const active = (state.history ?? []).filter((row) => !row.effectiveTo && row.specializationId && row.gradeId);
        setTracks(active.map((row) => ({ specializationId: row.specializationId as string, gradeId: row.gradeId as string })));
      })
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    if (!session?.expiresAt || result) return;
    const tick = () => setClock(left(session.expiresAt) ?? "");
    tick();
    const id = window.setInterval(tick, 1000);
    return () => window.clearInterval(id);
  }, [session?.expiresAt, result]);

  const items = session?.items ?? [];
  const item = items[index];
  const answered = items.filter((entry) => entry.status === "answered").length;
  const highlighted = useMemo(() => (item ? paint(item.body, item.type) : null), [item]);
  const held = tracks.find((track) => track.specializationId === specializationId);
  const heldGrade = byId(catalog?.grades, held?.gradeId);
  const grades = (catalog?.grades ?? []).filter((grade) => !heldGrade || (grade.rank ?? 0) > (heldGrade.rank ?? 0));

  useEffect(() => {
    const current = items[index];
    if (!current) return;
    const saved = answers[current.id];
    if (Array.isArray(saved)) {
      setChoice(saved);
      setText("");
      return;
    }
    if (typeof saved === "string" && (current.type === "single_choice" || current.type === "multi_choice") && current.options?.length) {
      setChoice(saved ? [saved] : []);
      setText("");
      return;
    }
    setChoice([]);
    setText(typeof saved === "string" ? saved : "");
  }, [index, session?.id]);

  function remember(nextIndex: number) {
    setConfirmFinish(false);
    setError("");
    setIndex(nextIndex);
  }

  async function start() {
    if (!catalog) return;
    const picked = byId(catalog.grades, gradeId);
    const specName = byId(catalog.specializations, specializationId)?.name ?? "этом направлении";
    if (heldGrade && picked && (picked.rank ?? 0) <= (heldGrade.rank ?? 0)) {
      setError(`В направлении «${specName}» уже ${heldGrade.name}. Тот же уровень и уровни ниже закрыты. Можно сдать только более высокий грейд.`);
      return;
    }
    const category = categoryFor(catalog, specializationId, gradeId);
    if (!category) {
      setError("Для этой пары специализации и грейда нет категории.");
      return;
    }
    setPending(true);
    setError("");
    try {
      const created = await api<TestSession>("testing", "/me/sessions", {
        method: "POST",
        body: JSON.stringify({ targetCategoryId: category.id, gradeId }),
      });
      const full = await api<TestSession>("testing", `/me/sessions/${created.id}`);
      setSession(full.items ? full : { ...created, items: full.items ?? [] });
      setAnswers({});
      setIndex(0);
      setResult(null);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Сессия не стартовала");
    } finally {
      setPending(false);
    }
  }

  async function save(move: boolean) {
    if (!session || !item) return false;
    const draft = draftOf(item, choice, text);
    if (!filled(draft)) {
      if (move) {
        if (index < items.length - 1) remember(index + 1);
        return true;
      }
      setError("Нужен ответ.");
      return false;
    }
    if (sameDraft(answers[item.id], draft) && item.status === "answered") {
      if (move && index < items.length - 1) remember(index + 1);
      return true;
    }
    setPending(true);
    setError("");
    try {
      await api("testing", `/me/sessions/${session.id}/answers`, {
        method: "POST",
        body: JSON.stringify({ itemId: item.id, answer: draft }),
      });
      setAnswers((current) => ({ ...current, [item.id]: draft }));
      setSession({
        ...session,
        items: items.map((entry) => (entry.id === item.id ? { ...entry, status: "answered" as const } : entry)),
      });
      setConfirmFinish(false);
      if (move && index < items.length - 1) setIndex(index + 1);
      return true;
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Ответ не принят");
      return false;
    } finally {
      setPending(false);
    }
  }

  async function finish() {
    if (!session) return;
    const missing = items.filter((entry) => entry.status !== "answered");
    const draft = item ? draftOf(item, choice, text) : "";
    const unsaved = item ? filled(draft) && !sameDraft(answers[item.id], draft) : false;
    if ((missing.length > 0 || unsaved) && !confirmFinish) {
      const numbers = missing.map((entry) => items.indexOf(entry) + 1);
      const extra = unsaved && item && !numbers.includes(index + 1) ? [index + 1] : [];
      const list = [...numbers, ...extra].sort((a, b) => a - b).join(", ");
      setError(`Без сохранённого ответа: ${list}. Цветом в списке отмечены только записанные. Нажмите «Завершить тест» ещё раз, чтобы сдать так.`);
      setConfirmFinish(true);
      return;
    }
    setPending(true);
    setError("");
    try {
      const next = await api<SessionResult>("testing", `/me/sessions/${session.id}/submit`, { method: "POST" });
      if (next.decision === "confirmed") {
        await api("candidate", "/me/category", {
          method: "PUT",
          body: JSON.stringify({
            categoryId: next.resultingCategoryId,
            gradeId: next.resultingGradeId,
            specializationId,
          }),
        });
        setTracks((current) => [
          ...current.filter((track) => track.specializationId !== specializationId),
          { specializationId, gradeId: next.resultingGradeId },
        ]);
      }
      const [nextFsp, nextCooldown] = await Promise.all([
        api<FspState>("candidate", "/me/fsp").catch(() => null),
        api<GradeChanges>("testing", "/me/category-changes").catch(() => null),
      ]);
      setResult(next);
      setFsp(nextFsp);
      setCooldown(nextCooldown);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Тест не завершился");
    } finally {
      setPending(false);
    }
  }

  if (result && catalog) {
    const grade = byId<Named>(catalog.grades, result.resultingGradeId);
    const category = catalog.categories.find((entry) => entry.id === result.resultingCategoryId);
    const spec = byId<Named>(catalog.specializations, category?.specializationId);
    const confirmed = result.decision === "confirmed";
    return (
      <div className="grid gap-6">
        <PageTitle
          title={confirmed ? "Грейд подтверждён" : "Грейд не изменился"}
          text={confirmed
            ? `Балл ${Math.round(result.score)}.`
            : `Балл ${Math.round(result.score)}. Текущий уровень сохранён: понизить его этим тестом нельзя.`}
        />
        <p className="font-display text-4xl leading-none">{spec?.name ?? "Категория"} / {grade?.name ?? heldGrade?.name ?? "грейд"}</p>
        <p>{fsp?.fspMemberId ? `ФСП ${fsp.fspMemberId}` : "Истории ФСП нет"}</p>
        <p className="text-dim">
          {cooldown?.canChangeAt
            ? `Следующая смена грейда не раньше ${formatWhen(cooldown.canChangeAt)}.`
            : "Кулдаун не назначен."}
        </p>
      </div>
    );
  }

  if (!session) {
    return (
      <div className="grid max-w-xl gap-5">
        <PageTitle title="Тест на грейд" text="Грейд считается отдельно для каждой специализации. Повтор того же уровня и сдача ниже закрыты. Можно только повысить." />
        {error ? <Notice>{error}</Notice> : null}
        {!catalog ? <Empty title="Справочник молчит" text="Сервис dict не ответил. Без него нельзя выбрать категорию." /> : null}
        <Field label="Специализация">
          <Select
            value={specializationId}
            onChange={(event) => {
              const nextId = event.target.value;
              setSpecializationId(nextId);
              const floor = tracks.find((track) => track.specializationId === nextId);
              const floorGrade = byId(catalog?.grades, floor?.gradeId);
              const picked = byId(catalog?.grades, gradeId);
              if (floorGrade && picked && (picked.rank ?? 0) <= (floorGrade.rank ?? 0)) setGradeId("");
            }}
          >
            <option value="">Выберите</option>
            {catalog?.specializations.map((entry) => <option key={entry.id} value={entry.id}>{entry.name}</option>)}
          </Select>
        </Field>
        {heldGrade && grades.length === 0 ? (
          <p className="text-sm text-dim">В этом направлении уже {heldGrade.name}. Выше сдавать некуда, тот же уровень закрыт.</p>
        ) : (
          <Field label="Грейд, на который претендуете" hint={heldGrade ? `В этом направлении уже ${heldGrade.name}. Тот же уровень и ниже скрыты.` : "В этом направлении грейд ещё не подтверждён. Доступны все уровни."}>
            <Select value={gradeId} onChange={(event) => setGradeId(event.target.value)}>
              <option value="">Выберите</option>
              {grades.map((entry) => <option key={entry.id} value={entry.id}>{entry.name}</option>)}
            </Select>
          </Field>
        )}
        <Button disabled={pending || !specializationId || !gradeId} onClick={() => void start()}>
          {pending ? "Готовим задания" : "Начать"}
        </Button>
      </div>
    );
  }

  if (!item) {
    return <Empty title="В сессии нет заданий" text="Сервис вернул пустой список. Завершить тест всё равно можно." />;
  }

  const choiceType = item.type === "single_choice" || item.type === "multi_choice";
  const draft = draftOf(item, choice, text);
  const stored = item.status === "answered" && sameDraft(answers[item.id], draft);

  return (
    <div className="grid gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <PageTitle title={`Вопрос ${index + 1} из ${items.length}`} text={item.topic} />
        <p className="font-display text-3xl text-salary">{clock || (session.expiresAt ? "0:00" : "Без лимита")}</p>
      </div>
      <nav aria-label="Список вопросов" className="grid gap-3">
        <ol className="flex flex-wrap gap-2">
          {items.map((entry, entryIndex) => {
            const saved = entry.status === "answered";
            const current = entryIndex === index;
            return (
              <li key={entry.id}>
                <button
                  type="button"
                  aria-current={current ? "true" : undefined}
                  aria-label={`Вопрос ${entryIndex + 1}${saved ? ", ответ сохранён" : ", без ответа"}`}
                  className={`size-11 border text-sm ${saved ? "border-accent bg-accent text-on-accent" : "border-line text-dim"} ${current ? "outline outline-2 outline-offset-2 outline-accent" : ""}`}
                  onClick={() => remember(entryIndex)}
                >
                  {entryIndex + 1}
                </button>
              </li>
            );
          })}
        </ol>
        <p className="text-sm text-dim">Цветом отмечены вопросы с сохранённым ответом. «Дальше» и номер вопроса ответ не записывают.</p>
      </nav>
      {error ? <Notice>{error}</Notice> : null}
      {highlighted ? (
        <pre className="overflow-x-auto border border-line bg-raised p-4 text-sm" dangerouslySetInnerHTML={{ __html: highlighted }} />
      ) : (
        <p className="whitespace-pre-wrap">{item.body}</p>
      )}
      {choiceType && item.options && item.options.length > 0 ? (
        <div className="grid gap-2">
          {item.options.map((option) => {
            const on = choice.includes(option.id);
            return (
              <button
                key={option.id}
                type="button"
                className={`min-h-11 border px-3 text-left ${on ? "border-accent bg-deep" : "border-line"}`}
                onClick={() => {
                  setConfirmFinish(false);
                  if (item.type === "single_choice") setChoice([option.id]);
                  else setChoice(on ? choice.filter((id) => id !== option.id) : [...choice, option.id]);
                }}
              >
                {option.label}
              </button>
            );
          })}
        </div>
      ) : (
        <Field label="Ответ">
          <TextArea value={text} onChange={(event) => { setConfirmFinish(false); setText(event.target.value); }} />
        </Field>
      )}
      <div className="flex flex-wrap gap-3">
        <Button disabled={pending || !filled(draft) || stored} onClick={() => void save(false)}>
          {stored ? "Ответ сохранён" : item.status === "answered" ? "Перезаписать ответ" : "Сохранить ответ"}
        </Button>
        {index > 0 ? <Button variant="ghost" onClick={() => remember(index - 1)}>Назад</Button> : null}
        {index < items.length - 1 ? <Button variant="ghost" onClick={() => remember(index + 1)}>Дальше</Button> : null}
        <Button variant="deep" disabled={pending} onClick={() => void finish()}>
          {confirmFinish ? "Всё равно завершить" : "Завершить тест"}
        </Button>
      </div>
      <p className="text-sm text-dim">Сохранено {answered} из {items.length}.</p>
    </div>
  );
}
