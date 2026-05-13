export type StatsResponse = {
  applies_today: number;
  invitations: number;
  views: number;
};

export type StatCard = {
  id: 'applies' | 'invitations' | 'views' | 'replyRate';
  label: string;
  value: number;
  suffix?: string;
  hint: string;
};

export type PipelineStage = {
  name: string;
  value: number;
};
