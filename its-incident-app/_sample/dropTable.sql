SELECT 'DROP TABLE IF EXISTS ' || tablename || ' CASCADE;'
FROM pg_tables
WHERE tableowner = 'postgres';


-- 実行まで実施
SELECT 'DROP TABLE IF EXISTS ' || tablename || ' CASCADE;'
FROM pg_tables
WHERE tableowner = 'postgres' \gexec