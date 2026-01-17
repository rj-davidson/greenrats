export interface SSEMessage<T = unknown> {
  event: string;
  data: T;
}

export interface LeaderboardUpdate {
  tournamentId: string;
  entries: LeaderboardEntry[];
  updatedAt: string;
}

export interface LeaderboardEntry {
  golferId: string;
  golferName: string;
  position: number;
  score: number;
  thru: string;
  earnings: number;
}

export interface TournamentStatusUpdate {
  tournamentId: string;
  status: "upcoming" | "active" | "completed";
  round: number;
  updatedAt: string;
}

export type SSEEventType = "leaderboard" | "tournament_status" | "heartbeat";
