import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import type { CategoryState, FspState, Named, Profile } from "../api/types";
import { Empty, Notice, PageTitle, Skeleton } from "../components/ui";
import { CandidateTasks } from "./CandidateTasks";
import { byId, formatWhen } from "../lib/format";
import { loadCatalog, type Catalog } from "../lib/catalog";

export function CandidateHome() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [category, setCategory] = useState<CategoryState | null>(null);
  const [fsp, setFsp] = useState<FspState | null>(null);
  const [catalog, setCatalog] = useState<Catalog | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let live = true;
    Promise.all([
      api<Profile>("candidate", "/me/profile"),
      api<CategoryState>("candidate", "/me/category"),
      api<FspState>("candidate", "/me/fsp"),
      loadCatalog().catch(() => null),
    ])
      .then(([nextProfile, nextCategory, nextFsp, nextCatalog]) => {
        if (!live) return;
        setProfile(nextProfile);
        setCategory(nextCategory);
        setFsp(nextFsp);
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

  const tracks = (category?.history ?? []).filter((row) => !row.effectiveTo && row.specializationId && row.gradeId);
  const achievements = fsp?.achievements ?? [];

  return (
    <div className="grid gap-8">
      <PageTitle title={profile?.firstName ? `${profile.firstName} ${profile.lastName ?? ""}`.trim() : "Кабинет"} />
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
        <Link className="underline" to="/candidate/invitations">Приглашения</Link>
        <Link className="underline" to="/candidate/settings">Настройки</Link>
      </div>
    </div>
  );
}
