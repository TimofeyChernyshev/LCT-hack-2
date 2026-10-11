import { useEffect, useState, type FormEvent } from "react";
import { api } from "../api/client";
import type { Application, CandidateCard, Named, Vacancy } from "../api/types";
import { Button, Empty, Field, Modal, Notice, PageTitle, Salary, Select, Skeleton, TextArea, TextInput } from "../components/ui";
import { loadCatalog } from "../lib/catalog";
import { formatWhen, inviteStatus } from "../lib/format";

export function EmployerVacanciesPage() {
  const [vacancies, setVacancies] = useState<Vacancy[]>([]);
  const [specs, setSpecs] = useState<Named[]>([]);
  const [grades, setGrades] = useState<Named[]>([]);
  const [selectedVacancy, setSelectedVacancy] = useState<Vacancy | null>(null);
  const [applications, setApplications] = useState<Application[]>([]);
  const [candidateCards, setCandidateCards] = useState<Record<string, CandidateCard>>({});
  const [loading, setLoading] = useState(true);
  const [loadingApps, setLoadingApps] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");

  const [form, setForm] = useState({
    title: "",
    description: "",
    specializationId: "",
    gradeId: "",
    salaryMin: "",
    salaryMax: "",
    workFormat: "remote",
  });

  async function loadVacancies() {
    try {
      const list = await api<Vacancy[]>("employer", "/me/vacancies");
      setVacancies(list || []);
      setError("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить вакансии");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadCatalog()
      .then((cat) => {
        setSpecs(cat.specializations);
        setGrades(cat.grades);
      })
      .catch(() => undefined);
    void loadVacancies();
  }, []);

  async function loadApplications(vacancy: Vacancy) {
    if (!vacancy.id) return;
    setSelectedVacancy(vacancy);
    setLoadingApps(true);
    try {
      const apps = await api<Application[]>("interaction", `/me/vacancies/${vacancy.id}/applications`);
      setApplications(apps || []);

      // Load candidate cards for accepted applications
      const cards: Record<string, CandidateCard> = {};
      for (const app of apps || []) {
        try {
          const card = await api<CandidateCard>("candidate", `/candidates/${app.candidateUserId}`);
          cards[app.candidateUserId] = card;
        } catch {
          // ignore
        }
      }
      setCandidateCards(cards);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить отклики");
    } finally {
      setLoadingApps(false);
    }
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    const min = Number(form.salaryMin);
    const max = Number(form.salaryMax);
    if (!Number.isFinite(min) || !Number.isFinite(max) || min <= 0 || min > max) {
      setError("Зарплатная вилка обязательна: от > 0, от не больше до (в рублях).");
      return;
    }
    setPending(true);
    setError("");
    try {
      await api<Vacancy>("employer", "/me/vacancies", {
        method: "POST",
        body: JSON.stringify({
          title: form.title.trim(),
          description: form.description.trim(),
          salaryMin: min,
          salaryMax: max,
          workFormat: form.workFormat,
          currency: "RUB",
        }),
      });
      setShowCreate(false);
      setForm({ title: "", description: "", specializationId: "", gradeId: "", salaryMin: "", salaryMax: "", workFormat: "remote" });
      setInfo("Вакансия создана.");
      await loadVacancies();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать вакансию");
    } finally {
      setPending(false);
    }
  }

  async function handlePublish(id: string) {
    setError("");
    try {
      await api("employer", `/me/vacancies/${id}/publish`, { method: "POST" });
      setInfo("Вакансия опубликована.");
      await loadVacancies();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось опубликовать вакансию");
    }
  }

  async function handleAppDecision(appId: string, decision: "accept" | "reject") {
    setError("");
    try {
      await api("interaction", `/me/applications/${appId}/${decision}`, { method: "POST" });
      if (selectedVacancy) {
        await loadApplications(selectedVacancy);
      }
      setInfo(decision === "accept" ? "Отклик принят. Контакты соискателя открыты!" : "Отклик отклонён.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось обновить статус отклика");
    }
  }

  return (
    <div className="grid gap-8">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <PageTitle
          title="Вакансии компании"
          text="Публикация прямых вакансий и обработка входящих откликов соискателей."
        />
        <Button onClick={() => setShowCreate(true)}>+ Новая вакансия</Button>
      </div>

      {error ? <Notice>{error}</Notice> : null}
      {info ? <Notice tone="ok">{info}</Notice> : null}

      {loading ? <Skeleton className="h-40" /> : null}

      {!loading && vacancies.length === 0 ? (
        <Empty
          title="Вакансий пока нет"
          text="Опубликуйте первую вакансию с обязательной зарплатной вилкой, чтобы соискатели могли откликаться."
        />
      ) : null}

      <div className="grid gap-6">
        {vacancies.map((v) => (
          <div key={v.id} className="border border-line bg-bg p-5">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <div className="flex items-center gap-3">
                  <h2 className="font-display text-2xl leading-none">{v.title}</h2>
                  <span className={`px-2 py-0.5 text-xs font-medium uppercase ${v.status === "published" ? "bg-accent/20 text-accent" : "bg-raised text-dim"}`}>
                    {v.status === "published" ? "Опубликована" : "Черновик"}
                  </span>
                </div>
                <p className="mt-2 text-sm text-dim">Создана {formatWhen(v.createdAt)}</p>
              </div>
              <Salary min={v.salaryMin} max={v.salaryMax} />
            </div>

            <p className="mt-4 max-w-[72ch] whitespace-pre-wrap text-sm">{v.description}</p>

            <div className="mt-5 flex flex-wrap items-center gap-3 border-t border-line pt-4">
              {v.status !== "published" && v.id ? (
                <Button variant="ghost" onClick={() => void handlePublish(v.id!)}>
                  Опубликовать
                </Button>
              ) : null}
              <Button variant="ghost" onClick={() => void loadApplications(v)}>
                Просмотреть отклики
              </Button>
            </div>
          </div>
        ))}
      </div>

      {/* Modal / Drawer for vacancy applications */}
      {selectedVacancy ? (
        <Modal
          title={`Отклики: ${selectedVacancy.title}`}
          onClose={() => setSelectedVacancy(null)}
        >
          <div className="grid gap-4">
            {loadingApps ? <Skeleton className="h-28" /> : null}

            {!loadingApps && applications.length === 0 ? (
              <Empty
                title="Откликов пока нет"
                text="Когда кандидаты откликнутся на эту вакансию, их заявки появятся здесь."
              />
            ) : null}

            <ul className="divide-y divide-line">
              {applications.map((app) => {
                const card = candidateCards[app.candidateUserId];
                const isOpen = app.status === "accepted";
                const isPending = app.status === "submitted" || app.status === "viewed";

                return (
                  <li key={app.id} className="py-4">
                    <div className="flex flex-wrap items-baseline justify-between gap-2">
                      <span className="font-medium text-ink">
                        {card?.displayName || "Соискатель"}
                      </span>
                      <span className="text-xs text-dim">
                        {inviteStatus[app.status] ?? app.status} · {formatWhen(app.createdAt)}
                      </span>
                    </div>

                    {app.coverLetter ? (
                      <p className="mt-2 text-sm text-dim">
                        <strong className="text-ink">Сопроводительное:</strong> {app.coverLetter}
                      </p>
                    ) : null}

                    {isOpen && card?.contacts ? (
                      <div className="mt-3 border border-line bg-deep/40 p-3 text-sm">
                        <p className="font-medium text-salary">Контакты открыты:</p>
                        <ul className="mt-1 grid gap-0.5 text-xs">
                          {card.contacts.email ? <li>Email: {card.contacts.email}</li> : null}
                          {card.contacts.phone ? <li>Телефон: {card.contacts.phone}</li> : null}
                          {card.contacts.telegram ? <li>Telegram: {card.contacts.telegram}</li> : null}
                        </ul>
                      </div>
                    ) : null}

                    {isPending ? (
                      <div className="mt-3 flex gap-2">
                        <Button className="text-xs" onClick={() => void handleAppDecision(app.id, "accept")}>
                          Принять и открыть контакты
                        </Button>
                        <Button variant="ghost" className="text-xs" onClick={() => void handleAppDecision(app.id, "reject")}>
                          Отклонить
                        </Button>
                      </div>
                    ) : null}
                  </li>
                );
              })}
            </ul>
          </div>
        </Modal>
      ) : null}

      {/* Modal for creating a new vacancy */}
      {showCreate ? (
        <Modal title="Создание вакансии" onClose={() => setShowCreate(false)}>
          <form className="grid gap-4" onSubmit={handleCreate}>
            <Field label="Название должности">
              <TextInput
                required
                placeholder="Senior Go Developer"
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
              />
            </Field>

            <div className="grid gap-3 sm:grid-cols-2">
              <Field label="Специализация">
                <Select value={form.specializationId} onChange={(e) => setForm({ ...form, specializationId: e.target.value })}>
                  <option value="">Любая</option>
                  {specs.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
                </Select>
              </Field>
              <Field label="Грейд">
                <Select value={form.gradeId} onChange={(e) => setForm({ ...form, gradeId: e.target.value })}>
                  <option value="">Любой</option>
                  {grades.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
                </Select>
              </Field>
            </div>

            <Field label="Описание роли и требований">
              <TextArea
                required
                rows={4}
                placeholder="Опишите стек, задачи команды и ожидания от кандидата..."
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
              />
            </Field>

            <Field label="Формат работы">
              <Select value={form.workFormat} onChange={(e) => setForm({ ...form, workFormat: e.target.value })}>
                <option value="remote">Удалённо</option>
                <option value="hybrid">Гибрид</option>
                <option value="office">Офис</option>
              </Select>
            </Field>

            <div className="grid gap-3 sm:grid-cols-2">
              <Field label="Вилка от, ₽" hint="Обязательно">
                <TextInput
                  required
                  inputMode="numeric"
                  placeholder="200000"
                  value={form.salaryMin}
                  onChange={(e) => setForm({ ...form, salaryMin: e.target.value })}
                />
              </Field>
              <Field label="Вилка до, ₽" hint="Обязательно">
                <TextInput
                  required
                  inputMode="numeric"
                  placeholder="300000"
                  value={form.salaryMax}
                  onChange={(e) => setForm({ ...form, salaryMax: e.target.value })}
                />
              </Field>
            </div>
            <p className="text-xs text-dim">В рублях, до вычета НДФЛ.</p>

            <div className="mt-2 flex gap-3">
              <Button type="submit" disabled={pending}>
                {pending ? "Создаём..." : "Создать вакансию"}
              </Button>
              <Button type="button" variant="ghost" onClick={() => setShowCreate(false)}>
                Отмена
              </Button>
            </div>
          </form>
        </Modal>
      ) : null}
    </div>
  );
}
