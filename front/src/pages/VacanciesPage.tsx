import { useEffect, useState } from "react";
import { api } from "../api/client";
import type { Application, Vacancy } from "../api/types";
import { Button, Empty, Field, Notice, PageTitle, Salary, Skeleton, TextArea } from "../components/ui";
import { inviteStatus } from "../lib/format";

export function VacanciesPage() {
  const [items, setItems] = useState<Vacancy[]>([]);
  const [mine, setMine] = useState<Application[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [letter, setLetter] = useState<Record<string, string>>({});
  const [pending, setPending] = useState("");

  useEffect(() => {
    let live = true;
    Promise.all([
      api<Vacancy[]>("employer", "/vacancies", { auth: false }),
      api<Application[]>("interaction", "/me/applications").catch(() => [] as Application[]),
    ])
      .then(([vacancies, applications]) => {
        if (!live) return;
        setItems(vacancies);
        setMine(applications);
      })
      .catch((reason: unknown) => {
        if (live) setError(reason instanceof Error ? reason.message : "Вакансии не загрузились");
      })
      .finally(() => {
        if (live) setLoading(false);
      });
    return () => {
      live = false;
    };
  }, []);

  async function apply(vacancy: Vacancy) {
    if (!vacancy.id) return;
    setPending(vacancy.id);
    setError("");
    try {
      const created = await api<Application>("interaction", "/me/applications", {
        method: "POST",
        body: JSON.stringify({ vacancyId: vacancy.id, coverLetter: letter[vacancy.id] || "" }),
      });
      setMine((current) => [created, ...current]);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Отклик не ушёл");
    } finally {
      setPending("");
    }
  }

  return (
    <div className="grid gap-8">
      <PageTitle title="Открытые вакансии" />
      {error ? <Notice>{error}</Notice> : null}
      {loading ? <Skeleton className="h-40" /> : null}
      {!loading && items.length === 0 ? <Empty title="Вакансий нет" text="Работодатели ещё ничего не опубликовали." /> : null}
      <ul className="grid gap-8">
        {items.map((item) => {
          const sent = mine.find((entry) => entry.vacancyId === item.id);
          return (
            <li key={item.id} className="border-t border-line pt-6">
              <div className="grid gap-4 md:grid-cols-[1fr_auto] md:items-start">
                <div>
                  <h2 className="font-display text-3xl leading-none">{item.title}</h2>
                  <p className="mt-3 max-w-[68ch] whitespace-pre-wrap text-dim">{item.description}</p>
                </div>
                <Salary min={item.salaryMin} max={item.salaryMax} />
              </div>
              {sent ? (
                <p className="mt-4 text-sm">Отклик: {inviteStatus[sent.status] ?? sent.status}</p>
              ) : (
                <div className="mt-4 grid max-w-xl gap-3">
                  <Field label="Сопроводительное">
                    <TextArea value={letter[item.id ?? ""] ?? ""} onChange={(event) => setLetter({ ...letter, [item.id ?? ""]: event.target.value })} />
                  </Field>
                  <Button disabled={pending === item.id} onClick={() => void apply(item)}>
                    {pending === item.id ? "Отправляем" : "Откликнуться"}
                  </Button>
                </div>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
