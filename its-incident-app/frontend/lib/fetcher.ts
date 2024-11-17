export const fetcher = async <T>(
  url: string,
  init?: RequestInit,
): Promise<T> => {
  const res = await fetch(url, init);

  if (!res.ok) {
    const error = new Error("APIエラーが発生しました");
    error.message = await res.text();
    throw error;
  }

  return res.json();
};
