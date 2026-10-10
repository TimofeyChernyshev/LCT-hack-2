import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from "react";
import logoOnDark from "../assets/brand/logo-short-on-dark.svg";
import logoOnLight from "../assets/brand/logo-short-on-light.svg";

export function Wordmark() {
  return (
    <span className="inline-flex items-center">
      <img src={logoOnDark} alt="ФСП" className="logo-on-dark h-8 w-auto" />
      <img src={logoOnLight} alt="" className="logo-on-light h-8 w-auto" />
    </span>
  );
}

export function PageTitle({ title, text }: { title: string; text?: string }) {
  return (
    <header className="mb-8 max-w-[68ch]">
      <h1 className="font-display text-4xl leading-[1.1] text-ink md:text-5xl">{title}</h1>
      {text ? <p className="mt-4 text-dim leading-relaxed">{text}</p> : null}
    </header>
  );
}

const buttonClass = {
  primary: "bg-accent text-on-accent hover:brightness-110",
  salary: "bg-salary text-on-accent hover:brightness-110",
  ghost: "border border-line bg-transparent text-ink hover:bg-raised",
  deep: "bg-deep text-ink hover:brightness-125",
};

export function Button({
  variant = "primary",
  className = "",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: keyof typeof buttonClass }) {
  return (
    <button
      className={`inline-flex min-h-11 items-center justify-center px-4 text-sm transition active:translate-y-px disabled:cursor-not-allowed disabled:opacity-50 ${buttonClass[variant]} ${className}`}
      {...props}
    />
  );
}

export function Field({
  label,
  error,
  hint,
  children,
}: {
  label: string;
  error?: string;
  hint?: string;
  children: ReactNode;
}) {
  return (
    <label className="grid gap-2 text-sm">
      <span>{label}</span>
      {children}
      {hint ? <span className="text-dim">{hint}</span> : null}
      {error ? <span className="text-salary">{error}</span> : null}
    </label>
  );
}

const control =
  "min-h-11 w-full border border-line bg-bg px-3 text-ink placeholder:text-dim";

export function TextInput({ className = "", ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={`${control} ${className}`} {...props} />;
}

export function TextArea({ className = "", ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className={`${control} min-h-28 py-3 ${className}`} {...props} />;
}

export function Select({ className = "", ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={`${control} ${className}`} {...props} />;
}

export function Notice({ children, tone = "error" }: { children: ReactNode; tone?: "error" | "ok" }) {
  return (
    <p className={`border border-line px-3 py-2 text-sm ${tone === "error" ? "text-salary" : "text-ink"}`} role={tone === "error" ? "alert" : "status"}>
      {children}
    </p>
  );
}

export function Empty({ title, text }: { title: string; text: string }) {
  return (
    <div className="border border-dashed border-line px-4 py-10">
      <p className="font-display text-2xl">{title}</p>
      <p className="mt-2 max-w-[52ch] text-dim">{text}</p>
    </div>
  );
}

export function Skeleton({ className = "" }: { className?: string }) {
  return <div className={`motion-safe:animate-pulse bg-raised ${className}`} />;
}

export function Modal({
  title,
  children,
  onClose,
}: {
  title: string;
  children: ReactNode;
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 z-40 flex items-end justify-center bg-black/50 p-4 sm:items-center" role="presentation" onClick={onClose}>
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="dialog-title"
        className="w-full max-w-lg border border-line bg-bg p-5"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="mb-4 flex items-start justify-between gap-4">
          <h2 id="dialog-title" className="font-display text-2xl leading-none">
            {title}
          </h2>
          <button type="button" className="min-h-11 px-2 text-dim" onClick={onClose}>
            Закрыть
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

export function Salary({ min, max }: { min: number; max: number }) {
  return (
    <p>
      <span className="font-display text-3xl leading-none text-salary md:text-4xl">
        от {new Intl.NumberFormat("ru-RU").format(min)} до {new Intl.NumberFormat("ru-RU").format(max)} ₽
      </span>
      <span className="mt-2 block text-sm text-dim">до вычета НДФЛ</span>
    </p>
  );
}
