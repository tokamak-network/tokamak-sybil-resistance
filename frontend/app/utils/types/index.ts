export interface LeaderboardEntry {
  rank: number;
  address: string;
  score: number;
  balance: number;
}

export interface WalletContextType {
  address: string | null;
  balance: string | null;
  connectWallet: () => Promise<void>;
}
