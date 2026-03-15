import { mockTerminalSessions } from "./mockSessions";
import { planFollowupActions, prioritizeActions, type FollowupAction } from "./followupPlanner";

export type SessionFollowupDeck = {
  sessionID: string;
  sessionTitle: string;
  actions: FollowupAction[];
};

export const mockFollowupDecks: SessionFollowupDeck[] = mockTerminalSessions.map((session) => ({
  sessionID: session.id,
  sessionTitle: session.title,
  actions: prioritizeActions(planFollowupActions(session))
}));
