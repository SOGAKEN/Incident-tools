// UserContext.tsx
import { createContext } from "react";

type HeaderData = {
  name: string;
  email: string;
  image: string;
  user_id: number;
};

// 初期値は null とします
export const UserContext = createContext<HeaderData | null>(null);
