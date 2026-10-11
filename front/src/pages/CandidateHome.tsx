import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, downloadFile } from "../api/client";
import type { CategoryState, FspState, Named, Profile, TestSession } from "../api/types";
import { Button, Empty, Notice, PageTitle, Skeleton } from "../components/ui";
import { CandidateTasks } from "./CandidateTasks";
import { byId, formatWhen } from "../lib/format";
import { loadCatalog, type Catalog } from "../lib/catalog";

export function CandidateHome() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [category, setCategory] = useState<CategoryState | null>(null);
  const [fsp, setFsp] = useState<FspState | null>(null);
  const [catalog, setCatalog] = useState<Catalog | null>(null);
  const [sessions, setSessions] = useState<TestSession[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [downloading, setDownloading] = useState(false);

  useEffect(() => {
    let live = true;
    Promise.all([
      api<Profile>("candidate", "/me/profile"),
      api<CategoryState>("candidate", "/me/category"),
      api<FspState>("candidate", "/me/fsp"),
      api<TestSession[]>("testing", "/me/sessions").catch(() => [] as TestSession[]),
      loadCatalog().catch(() => null),
    ])
      .then(([nextProfile, nextCategory, nextFsp, nextSessions, nextCatalog]) => {
        if (!live) return;
        setProfile(nextProfile);
        setCategory(nextCategory);
        setFsp(nextFsp);
        setSessions(nextSessions || []);
        setCatalog(nextCatalog);
      })
      .catch((reason: unknown) => {
        if (live) setError(reason instanceof Error ? reason.message : "Кабинет не загрузился");
      })
      .finally(() => {
        if (live) setLoading(false);
      });
    return () => {
      live = false;
    };
  }, []);

  async function handleDownloadPdf() {
    setDownloading(true);
    try {
      await downloadFile("testing", "/candidates/export-pdf", `fsp_profile_${profile?.lastName || "candidate"}.pdf`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось скачать PDF профиль");
    } finally {
      setDownloading(false);
    }
  }

  const tracks = (category?.history ?? []).filter((row) => !row.effectiveTo && row.specializationId && row.gradeId);
  const achievements = fsp?.achievements ?? [];

  return (
    <div className="grid gap-8">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <PageTitle title={profile?.firstName ? `${profile.firstName} ${profile.lastName ?? ""}`.trim() : "Кабинет"} />
        <Button variant="ghost" disabled={downloading} onClick={() => void handleDownloadPdf()}>
          {downloading ? "Генерируем PDF..." : "📄 Скачать PDF-профиль (ФСП)"}
        </Button>
      </div>
      {error ? <Notice>{error}</Notice> : null}
      {loading ? <Skeleton className="h-36" /> : null}
      <section className="grid gap-6 border-t border-line pt-6">
        <div>
          <p className="text-sm text-dim">Подтверждённые грейды</p>
          {tracks.length === 0 ? <p className="mt-2 font-display text-3xl leading-none">Не подтверждены</p> : (
            <ul className="mt-4 divide-y divide-line border-y border-line">
              {tracks.map((track) => (
                <li key={track.specializationId} className="flex flex-wrap items-baseline justify-between gap-3 py-4">
                  <span>{byId<Named>(catalog?.specializations, track.specializationId)?.name ?? "Специализация"}</span>
                  <span className="font-display text-2xl leading-none">{byId<Named>(catalog?.grades, track.gradeId)?.name ?? "Грейд"}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
        <div>
          <p className="text-sm text-dim">ФСП</p>
          <p className="mt-2 font-display text-3xl leading-none">
            {fsp?.fspMemberId ? fsp.fspMemberId : "Истории нет"}
          </p>
        </div>
      </section>
      <section>
        <div className="flex items-baseline justify-between gap-2">
          <h2 className="font-display text-2xl">История тестирований</h2>
          <Link className="text-sm underline" to="/candidate/test">Пройти тест</Link>
        </div>
        {sessions.length === 0 ? (
          <div className="mt-4">
            <Empty title="Тестов пока нет" text="Пройдите опрос и тестирование для объективного подтверждения грейда." />
          </div>
        ) : (
          <ul className="mt-4 divide-y divide-line border-y border-line">
            {sessions.map((sess) => {
              const cat = catalog?.categories.find((c) => c.id === sess.targetCategoryId);
              const specName = byId<Named>(catalog?.specializations, cat?.specializationId)?.name;
              const gradeName = byId<Named>(catalog?.grades, cat?.gradeId)?.name;
              const categoryTitle = specName && gradeName ? `${specName} (${gradeName})` : "Тестирование квалификации";
              return (
                <li key={sess.id} className="flex flex-wrap items-baseline justify-between gap-3 py-4">
                  <div>
                    <span className="font-medium">{categoryTitle}</span>
                    <span className="ml-3 text-sm text-dim">
                      {sess.status === "finished" ? "Завершён" : sess.status === "active" ? "В процессе" : sess.status}
                    </span>
                  </div>
                  <span className="text-sm text-dim">{formatWhen(sess.startedAt)}</span>
                </li>
              );
            })}
          </ul>
        )}
      </section>
      <CandidateTasks />
      <section>
        <h2 className="font-display text-2xl">Достижения</h2>
        {achievements.length === 0 ? (
          <div className="mt-4">
            <Empty title="Достижений нет" text="Они появятся после связи с реестром ФСП." />
          </div>
        ) : (
          <ul className="mt-4 divide-y divide-line border-y border-line">
            {achievements.map((item) => (
              <li key={item.id} className="flex flex-wrap items-baseline justify-between gap-3 py-4">
                <span>{item.eventName}</span>
                <span className="text-dim">{item.place ? `${item.place} место` : formatWhen(item.eventDate)}</span>
              </li>
            ))}
          </ul>
        )}
      </section>
      <div className="flex flex-wrap gap-4 text-sm">
        <Link className="underline" to="/candidate/onboarding">Заполнить анкету</Link>
        <Link className="underline" to="/candidate/test">Пройти тест</Link>
        <Link className="underline" to="/candidate/tasks">Регулярные задания</Link>
        <Link className="underline" to="/candidate/invitations">Приглашения</Link>
        <Link className="underline" to="/candidate/settings">Настройки</Link>
      </div>
    </div>
  );
}
