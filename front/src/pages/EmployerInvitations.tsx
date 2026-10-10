import { useEffect, useState } from "react";
import { api } from "../api/client";
import type { CandidateCard, Invitation } from "../api/types";
import { Empty, Notice, PageTitle, Salary, Skeleton } from "../components/ui";
import { formatWhen, inviteStatus } from "../lib/format";

const columns = [
  { status: "sent", title: "Отправлено", hint: "Кандидат ещё не открыл" },
  { status: "viewed", title: "Просмотрено", hint: "Ждём решение" },
  { status: "accepted", title: "Принято", hint: "Можно связываться" },
  { status: "rejected", title: "Отклонено", hint: "Без раскрытия контактов" },
] as const;

type Opened = {
  name: string;
  lines: string[];
};

export function EmployerInvitations() {
  const [items, setItems] = useState<Invitation[]>([]);
  const [opened, setOpened] = useState<Record<string, Opened>>({});
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let live = true;
    api<Invitation[]>("interaction", "/me/invitations")
      .then(async (list) => {
        if (!live) return;
        setItems(list);
        const accepted = list.filter((item) => item.status === "accepted");
        const entries = await Promise.all(
          accepted.map(async (item) => {
            try {
              const card = await api<CandidateCard>("candidate", `/candidates/${item.candidateUserId}`);
              const lines = [
                card.contacts?.email,
                card.contacts?.phone,
                card.contacts?.telegram,
              ].filter((value): value is string => Boolean(value));
              return [item.id, {
                name: card.displayName || "Кандидат",
                lines: card.contacts?.masked === false ? lines : [],
              }] as const;
            } catch {
              return [item.id, { name: "Кандидат", lines: [] }] as const;
            }
          }),
        );
        if (live) setOpened(Object.fromEntries(entries));
      })
      .catch((reason: unknown) => {
        if (live) setError(reason instanceof Error ? reason.message : "Приглашения не загрузились");
      })
      .finally(() => {
        if (live) setLoading(false);
      });
    return () => {
      live = false;
    };
  }, []);

  const closed = items.filter((item) => item.status === "withdrawn" || item.status === "expired");

  return (
    <div className="grid gap-6">
      <PageTitle title="Исходящие" />
      {error ? <Notice>{error}</Notice> : null}
      {loading ? <Skeleton className="h-40" /> : null}
      {!loading && items.length === 0 ? <Empty title="Приглашений нет" text="Откройте каталог и отправьте первое." /> : null}
      {items.length > 0 ? (
        <div className="grid items-start gap-3 lg:grid-cols-4">
          {columns.map((column) => {
            const columnItems = items.filter((item) => item.status === column.status);
            return (
              <section key={column.status} className="grid content-start gap-3 border border-line p-3">
                <header className="grid gap-1 border-b border-line pb-3">
                  <div className="flex items-baseline justify-between gap-2">
                    <h2 className="font-display text-2xl leading-none">{column.title}</h2>
                    <span className="text-sm text-dim">{columnItems.length}</span>
                  </div>
                  <p className="text-sm text-dim">{column.hint}</p>
                </header>
                {columnItems.length === 0 ? <p className="text-sm text-dim">Пусто</p> : null}
                <ul className="grid gap-3">
                  {columnItems.map((item) => {
                    const person = opened[item.id];
                    return (
                      <li key={item.id} className="grid gap-3 border border-line bg-bg p-3">
                        <p className="text-sm text-dim">{formatWhen(item.createdAt)}</p>
                        <Salary min={item.salaryMin} max={item.salaryMax} />
                        <p className="line-clamp-4 text-sm">{item.message}</p>
                        {column.status === "accepted" ? (
                          <div className="grid gap-2 border-t border-line pt-3">
                            <p className="font-display text-2xl leading-none">{person?.name ?? "Кандидат"}</p>
                            <p className="text-sm">Приглашение принято. Контакты открыты.</p>
                            {person && person.lines.length > 0 ? (
                              <ul className="grid gap-1 text-sm">
                                {person.lines.map((line) => <li key={line}>{line}</li>)}
                              </ul>
                            ) : (
                              <p className="text-sm text-dim">В анкете нет почты, телефона и Telegram.</p>
                            )}
                          </div>
                        ) : null}
                      </li>
                    );
                  })}
                </ul>
              </section>
            );
          })}
        </div>
      ) : null}
      {closed.length > 0 ? (
        <section>
          <h2 className="mb-3 text-sm text-dim">Закрытые</h2>
          <ul className="divide-y divide-line border-y border-line">
            {closed.map((item) => (
              <li key={item.id} className="py-3 text-sm">{inviteStatus[item.status]} · {formatWhen(item.createdAt)}</li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
