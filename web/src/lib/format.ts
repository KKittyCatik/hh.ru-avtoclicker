import type { RealtimeEventName } from '@/types/api';

const compactFormatter = new Intl.NumberFormat('ru-RU', {
  notation: 'compact',
  maximumFractionDigits: 1,
});

const dateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
  dateStyle: 'short',
  timeStyle: 'short',
});

export function formatCompactNumber(value: number): string {
  return compactFormatter.format(value);
}

export function formatDateTime(value: string): string {
  return dateTimeFormatter.format(new Date(value));
}

export function formatRelativeTime(value: string): string {
  const target = new Date(value).getTime();
  const diffMinutes = Math.round((target - Date.now()) / 60000);
  const formatter = new Intl.RelativeTimeFormat('ru-RU', { numeric: 'auto' });

  if (Math.abs(diffMinutes) < 60) {
    return formatter.format(diffMinutes, 'minute');
  }

  const diffHours = Math.round(diffMinutes / 60);
  if (Math.abs(diffHours) < 24) {
    return formatter.format(diffHours, 'hour');
  }

  const diffDays = Math.round(diffHours / 24);
  return formatter.format(diffDays, 'day');
}

export function formatEventLabel(event: RealtimeEventName): string {
  switch (event) {
    case 'apply_result':
      return 'Отклик обработан';
    case 'new_invitation':
      return 'Новое приглашение';
    case 'new_message':
      return 'Новое сообщение';
    case 'resume_published':
      return 'Резюме опубликовано';
    case 'limit_reached':
      return 'Дневной лимит достигнут';
    default:
      return event;
  }
}
