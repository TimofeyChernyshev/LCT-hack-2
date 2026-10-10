import { useEffect, useState } from "react";
import { api } from "../api/client";
import type { Invitation } from "../api/types";
import { Button, Empty, Modal, Notice, PageTitle, Salary, Skeleton } from "../components/ui";
import { channelLabel, formatWhen, inviteStatus } from "../lib/format";

export function CandidateInvitations() {
  const [items, setItems] = useState<Invitation[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [pendingId, setPendingId] = useState("");
  const [accepting, setAccepting] = useState<Invitation | null>(null);

  async function load() {
    setLoading(true);
    try {
      setItems(await api<Invitation[]>("interaction", "/me/invitations/incoming"));
      setError("");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Приглашения не загрузились");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function act(item: Invitation, kind: "accept" | "reject") {
    setPendingId(item.id);
    setError("");
    try {
      const next = await api<Invitation>("interaction", `/me/invitations/${item.id}/${kind}`, { method: "POST" });
      setItems((current) => current.map((entry) => (entry.id === next.id ? next : entry)));
      setAccepting(null);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Статус не изменился");
    } finally {
      setPendingId("");
    }
  }

  return (
    <div className="grid gap-6">
      <PageTitle title="Приглашения" text="Вилка видна сразу. Контакты компании тоже. Ваши контакты откроются только после принятия." />
      {error ? <Notice>{error}</Notice> : null}
      {loading ? <Skeleton className="h-40" /> : null}
      {!loading && items.length === 0 ? (
        <Empty title="Пока пусто" text="Когда работодатель отправит приглашение, оно появится здесь." />
      ) : null}
      <ul className="grid gap-8">
        {items.map((item) => {
          const company = item.companyName || "Название компании не пришло";
          const open = item.status === "sent" || item.status === "viewed";
          return (
            <li key={item.id} className="border-t border-line pt-6">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <p className="text-sm text-dim">{inviteStatus[item.status] ?? item.status} · {formatWhen(item.createdAt)}</p>
                  <h2 className="mt-2 font-display text-3xl leading-none">{company}</h2>
                </div>
                <Salary min={item.salaryMin} max={item.salaryMax} />
              </div>
              <p className="mt-4 max-w-[68ch] whitespace-pre-wrap">{item.message}</p>
              <p className="mt-3 text-sm text-dim">
                {channelLabel[item.contactChannel] ?? item.contactChannel}: {item.employerContact || "контакт не указан"}
              </p>
              {open ? (
                <div className="mt-4 flex gap-3">
                  <Button disabled={pendingId === item.id} onClick={() => setAccepting(item)}>Принять</Button>
                  <Button variant="ghost" disabled={pendingId === item.id} onClick={() => void act(item, "reject")}>Отклонить</Button>
                </div>
              ) : null}
            </li>
          );
        })}
      </ul>
      {accepting ? (
        <Modal title="Открыть контакты" onClose={() => setAccepting(null)}>
          <p>
            Принимая приглашение, вы открываете свои контакты компании {accepting.companyName || "этому работодателю"}.
          </p>
          <div className="mt-5 flex gap-3">
            <Button disabled={pendingId === accepting.id} onClick={() => void act(accepting, "accept")}>Принять</Button>
            <Button variant="ghost" onClick={() => setAccepting(null)}>Назад</Button>
          </div>
        </Modal>
      ) : null}
    </div>
  );
}
