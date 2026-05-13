export type ApiErrorResponse = {
  error?: string;
  message?: string;
};

export type WorkerStatusResponse = {
  status: string;
};

export type GeneratedReply = {
  text: string;
  quickReplyOptionId?: string;
  needsHumanInput: boolean;
  isBotFlow: boolean;
};

export type RealtimeEventName =
  | 'new_invitation'
  | 'new_message'
  | 'apply_result'
  | 'resume_published'
  | 'limit_reached';

export type RealtimeEvent<TPayload = unknown> = {
  event: RealtimeEventName;
  payload: TPayload;
};
