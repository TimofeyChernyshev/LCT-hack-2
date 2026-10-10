import { useState } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import { List, SignOut, X } from "@phosphor-icons/react";
import { useAuth } from "../auth/session";
import { Button, Wordmark } from "./ui";

const candidateLinks = [
  { to: "/candidate", label: "Кабинет", end: true },
  { to: "/candidate/onboarding", label: "Анкета" },
  { to: "/candidate/test", label: "Тест" },
  { to: "/candidate/invitations", label: "Приглашения" },
  { to: "/candidate/vacancies", label: "Вакансии" },
  { to: "/candidate/settings", label: "Настройки" },
];

const employerLinks = [
  { to: "/employer", label: "Компания", end: true },
  { to: "/employer/candidates", label: "Каталог" },
  { to: "/employer/invitations", label: "Приглашения" },
  { to: "/employer/tasks", label: "Задания" },
  { to: "/employer/settings", label: "Настройки" },
];

export function Shell() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const links = user?.role === "employer" ? employerLinks : user?.role === "candidate" ? candidateLinks : [];

  return (
    <div className="min-h-dvh">
      <a href="#content" className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:bg-accent focus:px-3 focus:py-2">
        К содержанию
      </a>
      <header className="sticky top-0 z-30 border-b border-line bg-bg">
        <div className="mx-auto flex h-16 max-w-6xl items-center gap-4 px-4">
          <NavLink to="/" className="shrink-0" onClick={() => setOpen(false)}>
            <Wordmark />
          </NavLink>
          <nav className="hidden min-w-0 flex-1 items-center gap-1 md:flex" aria-label="Разделы">
            {links.map((link) => (
              <NavLink
                key={link.to}
                to={link.to}
                end={link.end}
                className={({ isActive }) =>
                  `px-3 py-2 text-sm ${isActive ? "bg-raised text-ink" : "text-dim hover:text-ink"}`
                }
              >
                {link.label}
              </NavLink>
            ))}
          </nav>
          <div className="ml-auto flex items-center gap-2">
            {user ? (
              <>
                <span className="hidden max-w-48 truncate text-sm text-dim sm:inline">{user.email}</span>
                <Button
                  variant="ghost"
                  className="px-3"
                  onClick={() => {
                    void logout().then(() => navigate("/"));
                  }}
                >
                  <SignOut size={18} />
                  <span className="sr-only">Выйти</span>
                </Button>
              </>
            ) : (
              <div className="hidden items-center gap-2 md:flex">
                <Link to="/login?role=candidate" className="px-3 py-2 text-sm text-dim hover:text-ink">
                  Соискатель
                </Link>
                <Link to="/login?role=employer" className="px-3 py-2 text-sm text-dim hover:text-ink">
                  Работодатель
                </Link>
              </div>
            )}
            <button
              type="button"
              className="inline-flex min-h-11 min-w-11 items-center justify-center md:hidden"
              aria-expanded={open}
              aria-label={open ? "Закрыть меню" : "Открыть меню"}
              onClick={() => setOpen((value) => !value)}
            >
              {open ? <X size={22} /> : <List size={22} />}
            </button>
          </div>
        </div>
        {open ? (
          <nav className="border-t border-line px-4 py-3 md:hidden" aria-label="Разделы">
            <div className="grid gap-1">
              {links.length
                ? links.map((link) => (
                    <NavLink
                      key={link.to}
                      to={link.to}
                      end={link.end}
                      className="px-2 py-3"
                      onClick={() => setOpen(false)}
                    >
                      {link.label}
                    </NavLink>
                  ))
                : (
                    <>
                      <Link to="/login?role=candidate" className="px-2 py-3" onClick={() => setOpen(false)}>Соискатель</Link>
                      <Link to="/login?role=employer" className="px-2 py-3" onClick={() => setOpen(false)}>Работодатель</Link>
                    </>
                  )}
            </div>
          </nav>
        ) : null}
      </header>
      <main id="content" className="mx-auto w-full max-w-6xl px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
