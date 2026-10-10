import { useState, useEffect, type FormEvent, type ReactNode } from "react";
import { Link, Navigate, useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import type { Role, User } from "../api/types";
import { useAuth, roleMismatchMessage } from "../auth/session";
import { Button, Field, Notice, PageTitle, Skeleton, TextInput } from "../components/ui";

function roleFromQuery(value: string | null): Role {
  return value === "employer" ? "employer" : "candidate";
}

export function Home() {
  const { user, ready } = useAuth();
  if (!ready) return <Skeleton className="h-40" />;
  if (user?.role === "candidate") return <Navigate to="/candidate" replace />;
  if (user?.role === "employer") return <Navigate to="/employer" replace />;

  return (
    <section className="grid items-end gap-10 md:grid-cols-[1.4fr_0.8fr]">
      <PageTitle
        title="Работодатель пишет первым"
        text="Категория по тесту, вилка в рублях до разговора, контакты после принятия приглашения."
      />
      <div className="grid gap-3">
        <Link to="/login?role=candidate" className="border border-line px-4 py-5 hover:bg-raised">
          <span className="font-display text-2xl">Соискатель</span>
          <span className="mt-2 block text-sm text-dim">Анкета, тест и входящие приглашения</span>
        </Link>
        <Link to="/login?role=employer" className="border border-line bg-deep px-4 py-5 hover:brightness-125">
          <span className="font-display text-2xl">Работодатель</span>
          <span className="mt-2 block text-sm text-dim">Каталог, объяснение выдачи и оффер</span>
        </Link>
      </div>
    </section>
  );
}

export function RequireRole({ role, children }: { role: "candidate" | "employer"; children: ReactNode }) {
  const { user, ready } = useAuth();
  if (!ready) return <Skeleton className="h-40" />;
  if (!user) return <Navigate to={`/login?role=${role}`} replace />;
  if (user.role !== role) {
    const home = user.role === "employer" ? "/employer" : user.role === "candidate" ? "/candidate" : "/";
    return (
      <div className="mx-auto grid max-w-md gap-5">
        <PageTitle title={role === "employer" ? "Кабинет работодателя" : "Кабинет соискателя"} />
        <Notice>{roleMismatchMessage(user.role)}</Notice>
        {home !== "/" ? <Link to={home} className="text-sm underline">Перейти в свой кабинет</Link> : null}
      </div>
    );
  }
  return children;
}

export function LoginPage() {
  const [params] = useSearchParams();
  const role = roleFromQuery(params.get("role"));
  const { login, user } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  if (user && user.role !== role) {
    const home = user.role === "employer" ? "/employer" : "/candidate";
    return (
      <div className="mx-auto grid max-w-md gap-5">
        <PageTitle title={role === "employer" ? "Вход работодателя" : "Вход соискателя"} />
        <Notice>{roleMismatchMessage(user.role)}</Notice>
        <Link to={home} className="text-sm underline">Перейти в свой кабинет</Link>
      </div>
    );
  }

  if (user) return <Navigate to={user.role === "employer" ? "/employer" : "/candidate"} replace />;

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    setError("");
    try {
      const next = await login(email.trim(), password, role);
      navigate(next.role === "employer" ? "/employer" : "/candidate");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Не удалось войти");
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="mx-auto grid max-w-md gap-5" onSubmit={onSubmit}>
      <PageTitle title={role === "employer" ? "Вход работодателя" : "Вход соискателя"} />
      {error ? <Notice>{error}</Notice> : null}
      <Field label="Почта">
        <TextInput type="email" autoComplete="email" required value={email} onChange={(event) => setEmail(event.target.value)} />
      </Field>
      <Field label="Пароль">
        <TextInput type="password" autoComplete="current-password" required value={password} onChange={(event) => setPassword(event.target.value)} />
      </Field>
      <Button type="submit" disabled={pending}>
        {pending ? "Входим" : "Войти"}
      </Button>
      <p className="text-sm text-dim">
        Нет аккаунта? <Link to={`/register?role=${role}`} className="text-ink underline">Регистрация</Link>
      </p>
    </form>
  );
}

export function RegisterPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [role, setRole] = useState<Role>(roleFromQuery(params.get("role")));
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [pdn, setPdn] = useState(false);
  const [publication, setPublication] = useState(false);
  const [error, setError] = useState("");
  const [done, setDone] = useState(false);
  const [pending, setPending] = useState(false);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    if (!pdn || !publication) {
      setError("Нужны оба согласия: на обработку данных и на публикацию профиля.");
      return;
    }
    if (password.length < 8) {
      setError("Пароль не короче 8 символов.");
      return;
    }
    setPending(true);
    setError("");
    try {
      await api<User>("auth", "/auth/register", {
        method: "POST",
        auth: false,
        body: JSON.stringify({
          email: email.trim(),
          password,
          role,
          consentPdn: true,
          consentProfilePublication: true,
        }),
      });
      setDone(true);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Регистрация не прошла");
    } finally {
      setPending(false);
    }
  }

  async function resend() {
    setError("");
    try {
      await api("auth", "/auth/resend-confirmation", {
        method: "POST",
        auth: false,
        body: JSON.stringify({ email: email.trim() }),
      });
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Письмо не отправилось");
    }
  }

  if (done) {
    return (
      <div className="mx-auto grid max-w-md gap-4">
        <PageTitle title="Проверьте почту" text="Мы отправили ссылку для подтверждения адреса. После этого можно войти." />
        {error ? <Notice>{error}</Notice> : null}
        <Button variant="ghost" onClick={() => void resend()}>Отправить письмо ещё раз</Button>
        <Button onClick={() => navigate(`/confirm-email?email=${encodeURIComponent(email.trim())}`)}>У меня есть токен</Button>
      </div>
    );
  }

  return (
    <form className="mx-auto grid max-w-md gap-5" onSubmit={onSubmit}>
      <PageTitle title="Регистрация" text="Почта с подтверждением и два согласия по 152-ФЗ." />
      {error ? <Notice>{error}</Notice> : null}
      <fieldset className="grid gap-2">
        <legend className="text-sm">Роль</legend>
        <div className="grid grid-cols-2 gap-2">
          {(["candidate", "employer"] as const).map((item) => (
            <button
              key={item}
              type="button"
              className={`min-h-11 border px-3 ${role === item ? "border-accent bg-deep" : "border-line"}`}
              onClick={() => setRole(item)}
            >
              {item === "candidate" ? "Соискатель" : "Работодатель"}
            </button>
          ))}
        </div>
      </fieldset>
      <Field label="Почта">
        <TextInput type="email" required autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} />
      </Field>
      <Field label="Пароль" hint="Не короче 8 символов">
        <TextInput type="password" required minLength={8} autoComplete="new-password" value={password} onChange={(event) => setPassword(event.target.value)} />
      </Field>
      <label className="flex items-start gap-3 text-sm">
        <input className="mt-1 size-4" type="checkbox" checked={pdn} onChange={(event) => setPdn(event.target.checked)} />
        <span>Согласен на обработку персональных данных.</span>
      </label>
      <label className="flex items-start gap-3 text-sm">
        <input className="mt-1 size-4" type="checkbox" checked={publication} onChange={(event) => setPublication(event.target.checked)} />
        <span>Согласен на публикацию профиля работодателям в обезличенном виде до принятия приглашения.</span>
      </label>
      <Button type="submit" disabled={pending}>{pending ? "Создаём" : "Создать аккаунт"}</Button>
      <p className="text-sm text-dim">
        Уже есть аккаунт? <Link className="text-ink underline" to={`/login?role=${role}`}>Войти</Link>
      </p>
    </form>
  );
}

export function ConfirmPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const initialToken = params.get("token") ?? "";
  const [token, setToken] = useState(initialToken);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  const confirmWithToken = async (tok: string) => {
    if (!tok.trim()) return;
    setPending(true);
    setError("");
    try {
      await api("auth", "/auth/confirm-email", {
        method: "POST",
        auth: false,
        body: JSON.stringify({ token: tok.trim() }),
      });
      setConfirmed(true);
      setPending(false);
      const role = params.get("role");
      setTimeout(() => {
        navigate(role ? `/login?role=${role}` : "/login", { replace: true });
      }, 1500);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Токен не принят или уже использован");
      setPending(false);
    }
  };

  useEffect(() => {
    if (initialToken) {
      void confirmWithToken(initialToken);
    }
  }, [initialToken]);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    await confirmWithToken(token);
  }

  return (
    <form className="mx-auto grid max-w-md gap-5" onSubmit={onSubmit}>
      <PageTitle
        title="Подтверждение почты"
        text={confirmed ? "Почта успешно подтверждена! Перенаправляем на страницу входа..." : "Подтверждение адреса электронной почты."}
      />
      {confirmed ? (
        <Notice>Почта успешно подтверждена! Сейчас откроется страница входа.</Notice>
      ) : null}
      {error ? <Notice>{error}</Notice> : null}
      {!confirmed && (
        <>
          <Field label="Токен">
            <TextInput required value={token} onChange={(event) => setToken(event.target.value)} />
          </Field>
          <Button type="submit" disabled={pending}>{pending ? "Проверяем токен..." : "Подтвердить"}</Button>
        </>
      )}
      <Link to="/login" className="text-sm underline">Ко входу</Link>
    </form>
  );
}
