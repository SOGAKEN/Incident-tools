import { Loader2 } from "lucide-react";

export default function Loading() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-background">
      <Loader2 className="h-16 w-16 animate-spin text-primary" />
      <h2 className="mt-4 text-xl font-semibold text-foreground">
        読み込み中...
      </h2>
      <p className="mt-2 text-sm text-muted-foreground">
        データを取得しています。しばらくお待ちください。
      </p>
    </div>
  );
}
