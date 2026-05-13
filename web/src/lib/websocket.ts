import type { RealtimeEvent } from '@/types/api';

export type SocketState = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'disconnected';

type Listener = (event: RealtimeEvent | { type: 'pong' }) => void;
type StateListener = (state: SocketState) => void;

function resolveWebSocketUrl(): string {
  const configured = import.meta.env.VITE_WS_URL?.trim();
  if (configured) {
    return configured;
  }
  return 'ws://localhost:8080/ws';
}

export class RealtimeClient {
  private socket: WebSocket | null = null;
  private reconnectAttempts = 0;
  private heartbeatTimer: number | null = null;
  private reconnectTimer: number | null = null;
  private listeners = new Set<Listener>();
  private stateListeners = new Set<StateListener>();
  private state: SocketState = 'idle';

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  onStateChange(listener: StateListener): () => void {
    this.stateListeners.add(listener);
    listener(this.state);
    return () => this.stateListeners.delete(listener);
  }

  connect(): void {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.setState(this.reconnectAttempts > 0 ? 'reconnecting' : 'connecting');
    const socket = new WebSocket(resolveWebSocketUrl());
    this.socket = socket;

    socket.onopen = () => {
      this.reconnectAttempts = 0;
      this.setState('connected');
      this.startHeartbeat();
    };

    socket.onmessage = (messageEvent) => {
      try {
        const parsed = JSON.parse(messageEvent.data as string) as RealtimeEvent | { type: 'pong' };
        this.listeners.forEach((listener) => listener(parsed));
      } catch {
        // Ignore malformed payloads.
      }
    };

    socket.onclose = () => {
      this.cleanupHeartbeat();
      this.setState('disconnected');
      this.scheduleReconnect();
    };

    socket.onerror = () => {
      socket.close();
    };
  }

  disconnect(): void {
    this.cleanupHeartbeat();
    if (this.reconnectTimer) {
      window.clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.socket?.close();
    this.socket = null;
    this.setState('disconnected');
  }

  private startHeartbeat(): void {
    this.cleanupHeartbeat();
    this.heartbeatTimer = window.setInterval(() => {
      if (this.socket?.readyState === WebSocket.OPEN) {
        this.socket.send(JSON.stringify({ type: 'ping' }));
      }
    }, 20_000);
  }

  private cleanupHeartbeat(): void {
    if (this.heartbeatTimer) {
      window.clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      return;
    }
    const delay = Math.min(30_000, 1_000 * 2 ** this.reconnectAttempts);
    this.reconnectAttempts += 1;
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  private setState(state: SocketState): void {
    this.state = state;
    this.stateListeners.forEach((listener) => listener(state));
  }
}

export const realtimeClient = new RealtimeClient();
