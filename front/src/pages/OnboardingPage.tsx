import { useEffect, useState, type FormEvent } from "react";
import { UploadSimple } from "@phosphor-icons/react";
import { api } from "../api/client";
import type { CandidateTechnology, Experience, Named, Profile, Resume, Technology } from "../api/types";
import { Button, Field, Notice, PageTitle, Select, TextArea, TextInput } from "../components/ui";
import { loadCatalog } from "../lib/catalog";

type Step = 0 | 1 | 2;
type PlaceKind = "company" | "freelance";

type Place = {
  key: string;
  serverId?: string;
  kind: PlaceKind;
  company: string;
  position: string;
  work: string;
  startedAt: string;
  endedAt: string;
};

const freelanceCompany = "Фриланс";

const stackGroups = [
  ["language", "Языки"],
  ["frontend", "Фронтенд"],
  ["mobile", "Мобильная разработка"],
  ["backend", "Бэкенд"],
  ["db", "Базы данных"],
  ["data", "Данные"],
  ["infra", "Инфраструктура"],
  ["qa", "Тестирование"],
] as const;

function blankPlace(): Place {
  return { key: crypto.randomUUID(), kind: "company", company: "", position: "", work: "", startedAt: "", endedAt: "" };
}

function isBlank(place: Place) {
  return !place.company.trim() && !place.position.trim() && !place.work.trim() && !place.startedAt && !place.endedAt;
}

function dateOnly(value?: string | null) {
  if (!value || value.startsWith("0001")) return "";
  return value.slice(0, 10);
}

function placeFromApi(item: Experience): Place {
  const freelance = item.company === freelanceCompany;
  return {
    key: item.id,
    serverId: item.id,
    kind: freelance ? "freelance" : "company",
    company: freelance ? "" : item.company,
    position: freelance ? "" : item.position,
    work: freelance ? item.description || (item.position === freelanceCompany ? "" : item.position) : "",
    startedAt: dateOnly(item.startedAt),
    endedAt: dateOnly(item.endedAt),
  };
}

export function OnboardingPage() {
  const [step, setStep] = useState<Step>(0);
  const [specs, setSpecs] = useState<Named[]>([]);
  const [techs, setTechs] = useState<Technology[]>([]);
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [pending, setPending] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [drag, setDrag] = useState(false);
  const [hasExperience, setHasExperience] = useState(false);
  const [places, setPlaces] = useState<Place[]>([]);
  const [removedIds, setRemovedIds] = useState<string[]>([]);
  const [techQuery, setTechQuery] = useState("");
  const [form, setForm] = useState({
    firstName: "",
    lastName: "",
    location: "",
    about: "",
    yearsExperience: "",
    specializationId: "",
    salaryMin: "",
    salaryMax: "",
    stack: [] as string[],
    fspMemberId: "",
  });

  useEffect(() => {
    let live = true;
    loadCatalog()
      .then((catalog) => {
        if (!live) return;
        setSpecs(catalog.specializations);
        setTechs(catalog.technologies);
      })
      .catch((reason: unknown) => {
        if (live) setError(reason instanceof Error ? reason.message : "Справочники не загрузились");
      });
    Promise.all([
      api<Profile>("candidate", "/me/profile"),
      api<Experience[]>("candidate", "/me/experiences").catch(() => [] as Experience[]),
      api<CandidateTechnology[]>("candidate", "/me/technologies").catch(() => [] as CandidateTechnology[]),
    ])
      .then(([profile, experiences, technologies]) => {
        if (!live) return;
        const loaded = experiences.map(placeFromApi);
        setPlaces(loaded);
        setHasExperience(loaded.length > 0 || (profile.yearsExperience ?? 0) > 0);
        setForm((current) => ({
          ...current,
          firstName: profile.firstName ?? "",
          lastName: profile.lastName ?? "",
          location: profile.location ?? "",
          about: profile.about ?? "",
          yearsExperience: profile.yearsExperience?.toString() ?? "",
          salaryMin: profile.salaryMin?.toString() ?? "",
          salaryMax: profile.salaryMax?.toString() ?? "",
          specializationId: profile.specializationId ?? "",
          stack: technologies.map((item) => item.technologyId),
          fspMemberId: profile.fspMemberId ?? "",
        }));
      })
      .catch(() => undefined);
    return () => {
      live = false;
    };
  }, []);

  function patch(partial: Partial<typeof form>) {
    setForm((current) => ({ ...current, ...partial }));
  }

  function chooseExperience(next: boolean) {
    setHasExperience(next);
    if (!next) patch({ yearsExperience: "0" });
    else if (form.yearsExperience === "0") patch({ yearsExperience: "" });
  }

  function updatePlace(key: string, partial: Partial<Place>) {
    setPlaces((current) => current.map((place) => (place.key === key ? { ...place, ...partial } : place)));
  }

  function removePlace(place: Place) {
    if (place.serverId) setRemovedIds((current) => [...current, place.serverId as string]);
    setPlaces((current) => current.filter((item) => item.key !== place.key));
  }

  function toggleTech(id: string) {
    patch({
      stack: form.stack.includes(id) ? form.stack.filter((item) => item !== id) : [...form.stack, id],
    });
  }

  function validatePlaces() {
    if (!hasExperience) return "";
    for (const place of places) {
      if (isBlank(place)) continue;
      if (!place.startedAt) return place.kind === "freelance" ? "У фриланса нужна дата начала." : "У места работы нужны компания, должность и дата начала.";
      if (place.kind === "company" && (!place.company.trim() || !place.position.trim())) {
        return "У места работы нужны компания, должность и дата начала.";
      }
      if (place.endedAt && place.endedAt < place.startedAt) return "Дата окончания раньше даты начала.";
    }
    return "";
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (!form.firstName.trim() || !form.lastName.trim()) {
      setError("Нужны имя и фамилия.");
      setStep(0);
      return;
    }
    const salaryMin = Number(form.salaryMin);
    const salaryMax = Number(form.salaryMax);
    if (!Number.isFinite(salaryMin) || !Number.isFinite(salaryMax) || salaryMin > salaryMax) {
      setError("Ожидания: укажите вилку, где «от» не больше «до».");
      setStep(1);
      return;
    }
    const placeError = validatePlaces();
    if (placeError) {
      setError(placeError);
      setStep(1);
      return;
    }
    setPending(true);
    setError("");
    setInfo("");
    try {
      await api("candidate", "/me/profile", {
        method: "PATCH",
        body: JSON.stringify({
          firstName: form.firstName,
          lastName: form.lastName,
          location: form.location,
          about: form.about,
          yearsExperience: hasExperience ? (form.yearsExperience ? Number(form.yearsExperience) : undefined) : 0,
          salaryMin,
          salaryMax,
          salaryCurrency: "RUB",
          specializationId: form.specializationId || undefined,
        }),
      });
      const dropIds = hasExperience ? removedIds : [...removedIds, ...places.flatMap((place) => (place.serverId ? [place.serverId] : []))];
      for (const id of dropIds) {
        await api("candidate", `/me/experiences/${id}`, { method: "DELETE" });
      }
      setRemovedIds([]);
      const nextPlaces = hasExperience ? places.filter((place) => !isBlank(place)) : [];
      for (const place of nextPlaces) {
        if (place.serverId) continue;
        const created = await api<Experience>("candidate", "/me/experiences", {
          method: "POST",
          body: JSON.stringify({
            company: place.kind === "freelance" ? freelanceCompany : place.company.trim(),
            position: place.kind === "freelance" ? place.work.trim() || freelanceCompany : place.position.trim(),
            startedAt: place.startedAt,
            endedAt: place.endedAt || undefined,
            description: place.kind === "freelance" ? place.work.trim() || undefined : undefined,
          }),
        });
        place.serverId = created.id;
      }
      setPlaces(nextPlaces);
      await api("candidate", "/me/technologies", {
        method: "PUT",
        body: JSON.stringify(form.stack.map((technologyId) => ({ technologyId, level: 3 }))),
      });
      if (form.fspMemberId.trim()) {
        await api("candidate", "/me/fsp", {
          method: "PUT",
          body: JSON.stringify({ fspMemberId: form.fspMemberId.trim() }),
        });
        setInfo("Профиль сохранён, ФСП ID привязан.");
      } else {
        setInfo("Профиль сохранён.");
      }
      if (file) {
        const body = new FormData();
        body.append("file", file);
        const resume = await api<Resume>("candidate", "/me/resumes/upload", { method: "POST", body });
        setInfo(`PDF разобран: ${resume.title}.`);
      }
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Сохранение не прошло");
    } finally {
      setPending(false);
    }
  }

  const query = techQuery.trim().toLowerCase();
  const visibleTechs = techs.filter((item) => !query || item.name.toLowerCase().includes(query) || item.code.toLowerCase().includes(query));
  const knownCategories = new Set<string>(stackGroups.map(([code]) => code));

  return (
    <form className="grid gap-6" onSubmit={save}>
      <PageTitle title="Анкета" text="Данные профиля и резюме." />
      {error ? <Notice>{error}</Notice> : null}
      {info ? <Notice tone="ok">{info}</Notice> : null}
      <div className="flex gap-2 text-sm">
        {["Имя", "Опыт", "ФСП и PDF"].map((label, index) => (
          <button
            key={label}
            type="button"
            className={`min-h-11 px-3 ${step === index ? "bg-raised" : "text-dim"}`}
            onClick={() => setStep(index as Step)}
          >
            {label}
          </button>
        ))}
      </div>
      {step === 0 ? (
        <div className="grid gap-4 md:grid-cols-2">
          <Field label="Имя"><TextInput required value={form.firstName} onChange={(event) => patch({ firstName: event.target.value })} /></Field>
          <Field label="Фамилия"><TextInput required value={form.lastName} onChange={(event) => patch({ lastName: event.target.value })} /></Field>
          <Field label="Город"><TextInput value={form.location} onChange={(event) => patch({ location: event.target.value })} /></Field>
          <div className="md:col-span-2">
            <Field label="О себе"><TextArea value={form.about} onChange={(event) => patch({ about: event.target.value })} /></Field>
          </div>
        </div>
      ) : null}
      {step === 1 ? (
        <div className="grid gap-6">
          <div className="grid gap-4 md:grid-cols-2">
            <Field label="Специализация">
              <Select value={form.specializationId} onChange={(event) => patch({ specializationId: event.target.value })}>
                <option value="">Выберите</option>
                {specs.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </Select>
            </Field>
            <Field label="Лет опыта">
              <TextInput
                inputMode="decimal"
                disabled={!hasExperience}
                value={hasExperience ? form.yearsExperience : "0"}
                onChange={(event) => patch({ yearsExperience: event.target.value })}
              />
            </Field>
            <div className="grid gap-4 md:col-span-2 md:grid-cols-2">
              <Field label="Ожидания от, ₽">
                <TextInput required inputMode="numeric" value={form.salaryMin} onChange={(event) => patch({ salaryMin: event.target.value })} />
              </Field>
              <Field label="Ожидания до, ₽">
                <TextInput required inputMode="numeric" value={form.salaryMax} onChange={(event) => patch({ salaryMax: event.target.value })} />
              </Field>
              <p className="text-sm text-dim md:col-span-2">До вычета НДФЛ</p>
            </div>
          </div>

          <fieldset className="grid gap-4">
            <legend className="text-sm">Опыт работы</legend>
            <div className="flex flex-wrap gap-2" role="radiogroup" aria-label="Опыт работы">
              <button type="button" className={`min-h-11 border px-3 text-sm ${hasExperience ? "border-line" : "border-accent bg-deep"}`} aria-pressed={!hasExperience} onClick={() => chooseExperience(false)}>
                Нет опыта
              </button>
              <button type="button" className={`min-h-11 border px-3 text-sm ${hasExperience ? "border-accent bg-deep" : "border-line"}`} aria-pressed={hasExperience} onClick={() => chooseExperience(true)}>
                Есть опыт
              </button>
            </div>
            {hasExperience ? (
              <div className="grid gap-4">
                {places.map((place, index) => (
                  <div key={place.key} className="grid gap-4 border border-line p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div className="flex flex-wrap gap-2">
                        <button type="button" className={`min-h-11 border px-3 text-sm ${place.kind === "company" ? "border-accent bg-deep" : "border-line"}`} onClick={() => updatePlace(place.key, { kind: "company" })}>
                          В компании
                        </button>
                        <button type="button" className={`min-h-11 border px-3 text-sm ${place.kind === "freelance" ? "border-accent bg-deep" : "border-line"}`} onClick={() => updatePlace(place.key, { kind: "freelance" })}>
                          Фриланс
                        </button>
                      </div>
                      <button type="button" className="min-h-11 px-3 text-sm text-dim" onClick={() => removePlace(place)}>
                        Убрать {index + 1}
                      </button>
                    </div>
                    {place.kind === "company" ? (
                      <div className="grid gap-4 md:grid-cols-2">
                        <Field label="Компания"><TextInput value={place.company} onChange={(event) => updatePlace(place.key, { company: event.target.value })} /></Field>
                        <Field label="Должность"><TextInput value={place.position} onChange={(event) => updatePlace(place.key, { position: event.target.value })} /></Field>
                      </div>
                    ) : (
                      <Field label="Чем занимались" hint="Компания и должность для фриланса не нужны.">
                        <TextInput value={place.work} onChange={(event) => updatePlace(place.key, { work: event.target.value })} />
                      </Field>
                    )}
                    <div className="grid gap-4 md:grid-cols-2">
                      <Field label="Начало"><TextInput type="date" value={place.startedAt} onChange={(event) => updatePlace(place.key, { startedAt: event.target.value })} /></Field>
                      <Field label="Окончание"><TextInput type="date" value={place.endedAt} onChange={(event) => updatePlace(place.key, { endedAt: event.target.value })} /></Field>
                      <p className="text-sm text-dim md:col-span-2">Пусто, если ещё идёт.</p>
                    </div>
                  </div>
                ))}
                <Button type="button" variant="ghost" className="justify-self-start" onClick={() => setPlaces((current) => [...current, blankPlace()])}>
                  Добавить место
                </Button>
              </div>
            ) : (
              <p className="text-sm text-dim">Компании указывать не нужно.</p>
            )}
          </fieldset>

          <fieldset className="grid gap-4">
            <legend className="text-sm">Стек</legend>
            <Field label="Поиск">
              <TextInput value={techQuery} placeholder="Go, Kafka, Flutter" onChange={(event) => setTechQuery(event.target.value)} />
            </Field>
            {stackGroups.map(([code, label]) => {
              const items = visibleTechs.filter((item) => item.category === code);
              if (!items.length) return null;
              return (
                <div key={code} className="grid gap-2">
                  <p className="text-sm text-dim">{label}</p>
                  <div className="flex flex-wrap gap-2">
                    {items.map((item) => {
                      const on = form.stack.includes(item.id);
                      return (
                        <button key={item.id} type="button" aria-pressed={on} className={`min-h-11 border px-3 text-sm ${on ? "border-accent bg-deep" : "border-line"}`} onClick={() => toggleTech(item.id)}>
                          {item.name}
                        </button>
                      );
                    })}
                  </div>
                </div>
              );
            })}
            {visibleTechs.some((item) => !item.category || !knownCategories.has(item.category)) ? (
              <div className="flex flex-wrap gap-2">
                {visibleTechs.filter((item) => !item.category || !knownCategories.has(item.category)).map((item) => {
                  const on = form.stack.includes(item.id);
                  return (
                    <button key={item.id} type="button" aria-pressed={on} className={`min-h-11 border px-3 text-sm ${on ? "border-accent bg-deep" : "border-line"}`} onClick={() => toggleTech(item.id)}>
                      {item.name}
                    </button>
                  );
                })}
              </div>
            ) : null}
          </fieldset>
        </div>
      ) : null}
      {step === 2 ? (
        <div className="grid gap-4">
          <Field label="ФСП ID">
            <TextInput value={form.fspMemberId} onChange={(event) => patch({ fspMemberId: event.target.value })} />
          </Field>
          <label
            className={`grid min-h-40 cursor-pointer place-items-center gap-3 border border-dashed px-4 py-8 text-center ${drag ? "border-accent bg-deep" : "border-line bg-raised/40 hover:border-accent"}`}
            onDragOver={(event) => {
              event.preventDefault();
              setDrag(true);
            }}
            onDragLeave={() => setDrag(false)}
            onDrop={(event) => {
              event.preventDefault();
              setDrag(false);
              const next = event.dataTransfer.files[0];
              if (next) setFile(next);
            }}
          >
            <UploadSimple size={28} aria-hidden="true" />
            <span className="max-w-[40ch] font-display text-2xl leading-none">
              {file ? file.name : "Загрузить PDF"}
            </span>
            <span className="text-sm text-dim">
              {file ? "Нажмите, чтобы заменить файл" : "Нажмите на область или перетащите резюме сюда"}
            </span>
            <span className="inline-flex min-h-11 items-center bg-accent px-4 text-sm text-on-accent">
              {file ? "Заменить PDF" : "Выбрать файл"}
            </span>
            <input
              className="sr-only"
              type="file"
              accept="application/pdf,.pdf"
              onChange={(event) => setFile(event.target.files?.[0] ?? null)}
            />
          </label>
        </div>
      ) : null}
      <div className="flex flex-wrap gap-3">
        {step < 2 ? (
          <Button type="button" onClick={() => setStep((step + 1) as Step)}>Дальше</Button>
        ) : (
          <Button type="submit" disabled={pending}>{pending ? "Сохраняем" : "Сохранить анкету"}</Button>
        )}
      </div>
    </form>
  );
}
