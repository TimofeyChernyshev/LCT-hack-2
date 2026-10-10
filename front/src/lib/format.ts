export function money(value: number) {
  return new Intl.NumberFormat("ru-RU").format(value);
}

export function salaryLine(min: number, max: number) {
  return `от ${money(min)} до ${money(max)} ₽`;
}

export const inviteStatus: Record<string, string> = {
  sent: "Отправлено",
  viewed: "Просмотрено",
  accepted: "Принято",
  rejected: "Отклонено",
  withdrawn: "Отозвано",
  expired: "Истекло",
};

export const decisionLabel: Record<string, string> = {
  confirmed: "Грейд подтверждён",
  downgrade_offered: "Можно сдать уровень ниже",
  upgrade_offered: "Можно подтвердить уровень выше",
};

export const channelLabel: Record<string, string> = {
  email: "Почта",
  telegram: "Telegram",
  phone: "Телефон",
  platform: "На платформе",
};

export function byId<T extends { id: string }>(items: T[] | undefined, id?: string | null) {
  if (!id || !items) return undefined;
  return items.find((item) => item.id === id);
}

export function formatWhen(iso?: string | null) {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}
