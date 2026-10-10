import { useEffect, useState, type FormEvent } from "react";
import { api } from "../api/client";
import type { Company, Named, Need } from "../api/types";
import { Button, Field, Modal, Notice, PageTitle, Salary, Select, TextArea, TextInput } from "../components/ui";
import { loadCatalog } from "../lib/catalog";

const workFormatLabel: Record<string, string> = {
  remote: "Удалённо",
  hybrid: "Гибрид",
  office: "Офис",
};

function formatName(items: Named[], id?: string | null) {
  if (!id) return "Любая";
  return items.find((item) => item.id === id)?.name ?? "Любая";
}

export function EmployerHome() {
  const [company, setCompany] = useState({ name: "", description: "", industry: "", website: "", size: "" });
  const [need, setNeed] = useState({
    title: "",
    description: "",
    specializationId: "",
    gradeId: "",
    salaryMin: "",
    salaryMax: "",
    workFormat: "remote",
  });
  const [specs, setSpecs] = useState<Named[]>([]);
  const [grades, setGrades] = useState<Named[]>([]);
  const [needs, setNeeds] = useState<Need[]>([]);
  const [viewing, setViewing] = useState<Need | null>(null);
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [pending, setPending] = useState(false);

  useEffect(() => {
    loadCatalog()
      .then((catalog) => {
        setSpecs(catalog.specializations);
        setGrades(catalog.grades);
      })
      .catch(() => undefined);
    api<Company>("employer", "/me/company")
      .then((value) => setCompany({
        name: value.name ?? "",
        description: value.description ?? "",
        industry: value.industry ?? "",
        website: value.website ?? "",
        size: value.size ?? "",
      }))
      .catch(() => undefined);
    api<Need[]>("employer", "/me/needs")
      .then(setNeeds)
      .catch(() => undefined);
  }, []);

  async function saveCompany(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    setError("");
    try {
      await api("employer", "/me/company", { method: "PUT", body: JSON.stringify(company) });
      setInfo("Компания сохранена.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Компания не сохранилась");
    } finally {
      setPending(false);
    }
  }

  async function removeNeed(id: string) {
    setError("");
    try {
      await api("interaction", `/me/needs/${id}/responses`, { method: "DELETE" }).catch(() => undefined);
      await api("employer", `/me/needs/${id}`, { method: "DELETE" });
      setNeeds((current) => current.filter((item) => item.id !== id));
      setViewing(null);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Потребность не удалилась");
    }
  }

  async function saveNeed(event: FormEvent) {
    event.preventDefault();
    const salaryMin = Number(need.salaryMin);
    const salaryMax = Number(need.salaryMax);
    if (!Number.isFinite(salaryMin) || !Number.isFinite(salaryMax) || salaryMin > salaryMax) {
      setError("В потребности вилка обязательна, «от» не больше «до».");
      return;
    }
    setPending(true);
    setError("");
    try {
      const created = await api<Need>("employer", "/me/needs", {
        method: "POST",
        body: JSON.stringify({
          title: need.title,
          description: need.description,
          specializationId: need.specializationId || undefined,
          gradeId: need.gradeId || undefined,
          salaryMin,
          salaryMax,
          workFormat: need.workFormat,
          status: "active",
        }),
      });
      setNeeds((current) => [created, ...current]);
      setNeed({ title: "", description: "", specializationId: "", gradeId: "", salaryMin: "", salaryMax: "", workFormat: "remote" });
      setInfo("Потребность сохранена.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Потребность не сохранилась");
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="grid gap-12">
      <PageTitle title="Компания" text="Опишите, кто вам нужен. По этому описанию собирается подборка." />
      {error ? <Notice>{error}</Notice> : null}
      {info ? <Notice tone="ok">{info}</Notice> : null}
      <form className="grid gap-4 md:grid-cols-2" onSubmit={saveCompany}>
        <Field label="Название"><TextInput required value={company.name} onChange={(event) => setCompany({ ...company, name: event.target.value })} /></Field>
        <Field label="Сайт"><TextInput value={company.website} onChange={(event) => setCompany({ ...company, website: event.target.value })} /></Field>
        <div className="md:col-span-2">
          <Field label="Чем занимаетесь"><TextArea required value={company.description} onChange={(event) => setCompany({ ...company, description: event.target.value })} /></Field>
        </div>
        <Field label="Направление"><TextInput value={company.industry} onChange={(event) => setCompany({ ...company, industry: event.target.value })} /></Field>
        <Field label="Размер"><TextInput value={company.size} onChange={(event) => setCompany({ ...company, size: event.target.value })} /></Field>
        <Button type="submit" disabled={pending}>Сохранить компанию</Button>
      </form>
      <form className="grid gap-4 border-t border-line pt-8 md:grid-cols-2" onSubmit={saveNeed}>
        <h2 className="font-display text-3xl md:col-span-2">Потребность</h2>
        <Field label="Роль"><TextInput required value={need.title} onChange={(event) => setNeed({ ...need, title: event.target.value })} /></Field>
        <Field label="Формат">
          <Select value={need.workFormat} onChange={(event) => setNeed({ ...need, workFormat: event.target.value })}>
            <option value="remote">Удалённо</option>
            <option value="hybrid">Гибрид</option>
            <option value="office">Офис</option>
          </Select>
        </Field>
        <div className="md:col-span-2">
          <Field label="Задачи команды"><TextArea required value={need.description} onChange={(event) => setNeed({ ...need, description: event.target.value })} /></Field>
        </div>
        <Field label="Специализация">
          <Select value={need.specializationId} onChange={(event) => setNeed({ ...need, specializationId: event.target.value })}>
            <option value="">Любая</option>
            {specs.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </Select>
        </Field>
        <Field label="Грейд">
          <Select value={need.gradeId} onChange={(event) => setNeed({ ...need, gradeId: event.target.value })}>
            <option value="">Любой</option>
            {grades.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
          </Select>
        </Field>
        <div className="grid gap-4 md:col-span-2 md:grid-cols-2">
          <Field label="Вилка от, ₽">
            <TextInput required inputMode="numeric" value={need.salaryMin} onChange={(event) => setNeed({ ...need, salaryMin: event.target.value })} />
          </Field>
          <Field label="Вилка до, ₽">
            <TextInput required inputMode="numeric" value={need.salaryMax} onChange={(event) => setNeed({ ...need, salaryMax: event.target.value })} />
          </Field>
          <p className="text-sm text-dim md:col-span-2">До вычета НДФЛ</p>
        </div>
        <Button className="md:col-span-2 md:justify-self-start" type="submit" disabled={pending}>Сохранить потребность</Button>
      </form>
      {needs.length > 0 ? (
        <ul className="divide-y divide-line border-y border-line">
          {needs.map((item) => (
            <li key={item.id}>
              <button type="button" className="flex w-full items-center justify-between gap-3 py-3 text-left" onClick={() => setViewing(item)}>
                <span>{item.title}</span>
                <span className="text-sm text-dim">Открыть</span>
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      {viewing ? (
        <Modal title={viewing.title} onClose={() => setViewing(null)}>
          <div className="grid gap-4">
            <p className="whitespace-pre-wrap">{viewing.description}</p>
            <p>{formatName(specs, viewing.specializationId) } · {formatName(grades, viewing.gradeId)}</p>
            <p>{workFormatLabel[viewing.workFormat ?? ""] ?? viewing.workFormat}</p>
            <Salary min={viewing.salaryMin} max={viewing.salaryMax} />
            <Button type="button" variant="ghost" onClick={() => void removeNeed(viewing.id)}>Удалить</Button>
          </div>
        </Modal>
      ) : null}
    </div>
  );
}
