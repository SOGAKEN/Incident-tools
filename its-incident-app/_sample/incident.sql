INSERT INTO incidents (
    datetime,
    status,
    judgment,
    content,
    assignee,
    priority,
    from_email,
    to_email,
    sender,
    subject
) VALUES (
    '2024-10-15 14:30:00',
    '調査中',
    '静観',
    'データベースサーバーへの接続が突然失われました。アプリケーションがデータを取得・保存できない状態です。緊急対応が必要です。データベースサーバーへの接続が突然失われました。アプリケーションがデータを取得・保存できない状態です。緊急対応が必要です。データベースサーバーへの接続が突然失われました。アプリケーションがデータを取得・保存できない状態です。緊急対応が必要です。データベースサーバーへの接続が突然失われました。アプリケーションがデータを取得・保存できない状態です。緊急対応が必要です。',
    '',
    '高',
    'alert@system.incidenttools.com',
    'support@incidenttools.com',
    'CTC',
    '【緊急】データベース接続エラー'
);