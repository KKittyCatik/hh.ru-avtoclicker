export type NegotiationStatus = 'active' | 'interview' | 'offer' | 'archived' | string;

export type NegotiationMessageOption = {
  id: string;
  text: string;
};

export type NegotiationMessage = {
  id: string;
  negotiation_id: string;
  type: string;
  from: string;
  text: string;
  options: NegotiationMessageOption[];
  created_at: string;
  negotiation_created_at: string;
  quick_reply_option_id?: string;
  needs_human_input?: boolean;
  potential_bot?: boolean;
};

export type Negotiation = {
  id: string;
  vacancy_id: string;
  resume_id: string;
  status: NegotiationStatus;
  is_bot: boolean;
  needs_reply: boolean;
  created_at: string;
  updated_at: string;
  last_message: NegotiationMessage;
};
