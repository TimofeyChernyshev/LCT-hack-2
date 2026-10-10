import { useEffect, useState, type FormEvent } from "react";
import { api } from "../api/client";
import type { FspState, Visibility } from "../api/types";
import { useAuth } from "../auth/session";
import { Button, Field, Notice, PageTitle, TextInput } from "../components/ui";

type Theme = "dark" | "light";

const themeKey = "fsp_theme";

function currentTheme(): Theme {
  return document.documentElement.dataset.theme === "light" ? "light" : "dark";
}

function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme;
  localStorage.setItem(themeKey, theme);
}

const privacyFields: { key: keyof Visibility; label: string }[] = [
  { key: "contacts", label: "Контакты" },
  { key: "links", label: "Ссылки" },
  { key: "fsp", label: "Достижения ФСП" },
  { key: "experience", label: "Опыт" },
  { key: "resume", label: "Резюме" },
  { key: "salary", label: "Зарплатные ожидания" },
  { key: "softSkills", label: "Софт-скиллы" },
];

export function SettingsPage() {
  const { user } = useAuth();
  const candidate = user?.role === "candidate";
  const [theme, setTheme] = useState<Theme>(currentTheme);
  const [visibility, setVisibility] = useState<Visibility>({});
  const [fspId, setFspId] = useState("");
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (!candidate) return;
    api<Visibility>("candidate", "/me/visibility")
      .then(setVisibility)
      .catch(() => undefined);
    api<FspState>("candidate", "/me/fsp")
      .then((state) => setFspId(state.fspMemberId ?? ""))
      .catch(() => undefined);
  }, [candidate]);

  function chooseTheme(next: Theme) {
    setTheme(next);
    applyTheme(next);
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (!candidate) return;
    setPending(true);
    setError("");
    setInfo("");
    try {
      await api("candidate", "/me/visibility", { method: "PUT", body: JSON.stringify(visibility) });
      if (fspId.trim()) {
        await api("candidate", "/me/fsp", { method: "PUT", body: JSON.stringify({ fspMemberId: fspId.trim() }) });
      }
      setInfo("Настройки сохранены.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Настройки не сохранились");
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="grid max-w-xl gap-8" onSubmit={save}>
      <PageTitle title="Настройки" />
      {error ? <Notice>{error}</Notice> : null}
      {info ? <Notice tone="ok">{info}</Notice> : null}
      <fieldset className="grid gap-3">
        <legend className="text-sm">Тема</legend>
        <div className="flex flex-wrap gap-2">
          <button type="button" aria-pressed={theme === "dark"} className={`min-h-11 border px-3 text-sm ${theme === "dark" ? "border-accent bg-deep" : "border-line"}`} onClick={() => chooseTheme("dark")}>
            Тёмная
          </button>
          <button type="button" aria-pressed={theme === "light"} className={`min-h-11 border px-3 text-sm ${theme === "light" ? "border-accent bg-deep" : "border-line"}`} onClick={() => chooseTheme("light")}>
            Светлая
          </button>
        </div>
        <p className="text-sm text-dim">Тёмная и светлая темы используют цвета брендбука.</p>
      </fieldset>
      {candidate ? (
        <>
          <fieldset className="grid gap-3">
            <legend className="text-sm">Что видит работодатель</legend>
            {privacyFields.map((field) => (
              <label key={field.key} className="flex min-h-11 items-center gap-3 text-sm">
                <input
                  type="checkbox"
                  checked={Boolean(visibility[field.key])}
                  onChange={(event) => setVisibility((current) => ({ ...current, [field.key]: event.target.checked }))}
                />
                {field.label}
              </label>
            ))}
          </fieldset>
          <Field label="ID участника ФСП">
            <TextInput value={fspId} onChange={(event) => setFspId(event.target.value)} />
          </Field>
          <p className="text-sm text-dim">Согласия на обработку и публикацию профиля приняты при регистрации.</p>
          <Button type="submit" disabled={pending}>{pending ? "Сохраняем" : "Сохранить настройки"}</Button>
        </>
      ) : (
        <p className="text-sm text-dim">Тема оформления.</p>
      )}
    </form>
  );
}
