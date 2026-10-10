import { useEffect, useState, type FormEvent } from "react";
import { api } from "../api/client";
import type { CategoryState, PeriodicTask } from "../api/types";
import { Button, Empty, Field, Notice, TextArea } from "../components/ui";
import { formatWhen } from "../lib/format";

export function CandidateTasks() {
  const [tasks, setTasks] = useState<PeriodicTask[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [pending, setPending] = useState("");

  async function load() {
    const [nextTasks, state] = await Promise.all([
      api<PeriodicTask[]>("testing", "/me/periodic-tasks"),
      api<CategoryState>("candidate", "/me/category").catch(() => null),
    ]);
    setTasks(nextTasks);
    setCategories((state?.history ?? []).filter((row) => !row.effectiveTo && row.categoryId).map((row) => row.categoryId as string));
  }

  useEffect(() => {
    load().catch((reason: unknown) => {
      setError(reason instanceof Error ? reason.message : "Задания не загрузились");
    });
  }, []);

  async function send(event: FormEvent<HTMLFormElement>, task: PeriodicTask) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const answer = String(form.get("answer") ?? "").trim();
    const kind = String(form.get("kind") ?? "solution") === "approach" ? "approach" : "solution";
    if (!answer) {
      setError("Нужен текст ответа.");
      return;
    }
    setPending(task.id);
    setError("");
    try {
      await api("testing", `/me/periodic-tasks/${task.id}/submit`, {
        method: "POST",
        body: JSON.stringify({ answer, kind }),
      });
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Ответ не отправился");
    } finally {
      setPending("");
    }
  }

  const visible = tasks.filter((task) => categories.includes(task.categoryId));
  const reactionLabel = {
    confirmed: "Работодатель подтвердил ответ.",
    thanks: "Работодатель поблагодарил.",
    invited: "Работодатель пригласил на работу.",
  };

  return (
    <section className="grid gap-4">
      <h2 className="font-display text-2xl">Задания от работодателей</h2>
      {error ? <Notice>{error}</Notice> : null}
      {visible.length === 0 ? <Empty title="Заданий нет" text="Их получают соискатели с подтверждённым грейдом в той специализации, для которой работодатель написал задачу." /> : null}
      <ul className="grid gap-6">
        {visible.map((task) => (
          <li key={task.id} className="border-t border-line pt-4">
            <div className="flex flex-wrap items-baseline justify-between gap-3">
              <h3 className="font-display text-2xl leading-none">{task.title}</h3>
              <span className="text-sm text-dim">{task.status === "sent" ? "Отправлено" : "Ждёт ответа"}</span>
            </div>
            <p className="mt-3 max-w-[68ch] whitespace-pre-wrap">{task.body}</p>
            {task.status === "sent" ? (
              <p className="mt-3 text-sm text-dim">
                {task.kind === "approach" ? "Подход" : "Решение"} · {formatWhen(task.submittedAt)}
                <span className="mt-1 block text-ink">{task.answer}</span>
                {task.reaction ? <span className="mt-2 block text-ink">{reactionLabel[task.reaction]}</span> : null}
              </p>
            ) : (
              <form className="mt-4 grid gap-3" onSubmit={(event) => void send(event, task)}>
                <div className="flex flex-wrap gap-2">
                  <label className="flex min-h-11 items-center gap-2 text-sm">
                    <input type="radio" name="kind" value="solution" defaultChecked />
                    Решить
                  </label>
                  <label className="flex min-h-11 items-center gap-2 text-sm">
                    <input type="radio" name="kind" value="approach" />
                    Предложить подход
                  </label>
                </div>
                <Field label="Ответ">
                  <TextArea name="answer" required />
                </Field>
                <Button type="submit" disabled={pending === task.id}>{pending === task.id ? "Отправляем" : "Отправить"}</Button>
              </form>
            )}
          </li>
        ))}
      </ul>
    </section>
  );
}
