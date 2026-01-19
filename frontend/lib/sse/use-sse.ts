"use client";

import type { SSEMessage, SSEEventType } from "@/lib/sse/types";
import { useEffect, useRef, useState } from "react";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

interface UseSSEOptions {
  topic: string;
  onMessage?: (message: SSEMessage) => void;
  onError?: (error: Event) => void;
  enabled?: boolean;
}

interface UseSSEReturn {
  isConnected: boolean;
  lastMessage: SSEMessage | null;
  error: Event | null;
}

export function useSSE({ topic, onMessage, onError, enabled = true }: UseSSEOptions): UseSSEReturn {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<SSEMessage | null>(null);
  const [error, setError] = useState<Event | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const connectRef = useRef<() => void>(() => {});

  useEffect(() => {
    const connect = () => {
      if (!enabled) return;

      const url = `${API_BASE_URL}/api/v1/sse?topic=${encodeURIComponent(topic)}`;
      const eventSource = new EventSource(url, { withCredentials: true });
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        setIsConnected(true);
        setError(null);
      };

      eventSource.onmessage = (event) => {
        try {
          const message: SSEMessage = JSON.parse(event.data);
          setLastMessage(message);
          onMessage?.(message);
        } catch {
          // Failed to parse SSE message
        }
      };

      eventSource.onerror = (err) => {
        setIsConnected(false);
        setError(err);
        onError?.(err);

        eventSource.close();
        reconnectTimeoutRef.current = setTimeout(() => {
          connectRef.current();
        }, 5000);
      };
    };

    connectRef.current = connect;
    connect();

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [topic, onMessage, onError, enabled]);

  return { isConnected, lastMessage, error };
}

export function useTournamentSSE(tournamentId: string, onUpdate?: (message: SSEMessage) => void) {
  return useSSE({
    topic: `tournament:${tournamentId}`,
    onMessage: onUpdate,
    enabled: !!tournamentId,
  });
}

export type { SSEMessage, SSEEventType };
