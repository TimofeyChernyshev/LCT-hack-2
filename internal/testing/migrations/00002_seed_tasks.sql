-- +goose Up
-- Базовые шаблоны тестирования
INSERT INTO test_templates (id, category_id, version, config, is_active)
VALUES
  ('00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', 1, '{"items_per_session": 10, "duration_minutes": 45, "min_score_to_pass": 0.7, "min_score_to_upgrade": 0.9}'::jsonb, TRUE),
  ('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 1, '{"items_per_session": 10, "duration_minutes": 45, "min_score_to_pass": 0.7, "min_score_to_upgrade": 0.9}'::jsonb, TRUE),
  ('22222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222', 1, '{"items_per_session": 10, "duration_minutes": 45, "min_score_to_pass": 0.7, "min_score_to_upgrade": 0.9}'::jsonb, TRUE)
ON CONFLICT (id) DO NOTHING;

-- Банк задач с поддержкой параметризации, античита (рандомизация опций/чисел) и уровней сложности (Junior=1-2, Middle=3-4, Senior=5)

-- 1. Junior: Алгоритмы и сложность (Single Choice, с перемешиванием)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000000',
  'single_choice',
  'algorithms',
  'Асимптотическая сложность бинарного поиска',
  'Какова временная сложность поиска элемента в отсортированном массиве из N элементов методом бинарного поиска в худшем случае?',
  '{}'::jsonb,
  '{
    "options": [
      {"id": "opt_log", "text": "O(log N)"},
      {"id": "opt_n", "text": "O(N)"},
      {"id": "opt_1", "text": "O(1)"},
      {"id": "opt_nlogn", "text": "O(N log N)"}
    ],
    "correct": ["opt_log"]
  }'::jsonb,
  1, 1.2, -1.5
) ON CONFLICT (id) DO NOTHING;

-- 2. Junior: Базы данных - индексы (Single Choice)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000000',
  'single_choice',
  'databases',
  'Тип индекса по умолчанию в PostgreSQL',
  'Какой тип индекса создается по умолчанию при вызове команды `CREATE INDEX` в PostgreSQL?',
  '{}'::jsonb,
  '{
    "options": [
      {"id": "opt_btree", "text": "B-Tree"},
      {"id": "opt_hash", "text": "Hash"},
      {"id": "opt_gin", "text": "GIN"},
      {"id": "opt_gist", "text": "GiST"}
    ],
    "correct": ["opt_btree"]
  }'::jsonb,
  2, 1.0, -1.0
) ON CONFLICT (id) DO NOTHING;

-- 3. Junior: Параметризованная задача на слайсы в Go (Античит: случайные параметры генератора)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000003',
  '00000000-0000-0000-0000-000000000000',
  'text',
  'golang',
  'Вместимость и длина слайса в Go',
  'Рассмотрите фрагмент кода на языке Go:
```go
s := make([]int, {{len_val}}, {{cap_val}})
s = append(s, {{add_val}})
println(cap(s))
```
Чему равна емкость `cap(s)` после выполнения этого фрагмента кода?',
  '{
    "variables": {
      "len_val": {"type": "int_range", "min": 2, "max": 4},
      "cap_val": {"type": "choice", "choices": ["8", "10", "12"]},
      "add_val": {"type": "int_range", "min": 100, "max": 999}
    },
    "formula": "cap_val"
  }'::jsonb,
  '{"expected": "{{cap_val}}"}'::jsonb,
  2, 1.1, -0.8
) ON CONFLICT (id) DO NOTHING;

-- 4. Middle: Многопоточность и гонки данных (Multi Choice)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000004',
  '00000000-0000-0000-0000-000000000000',
  'multi_choice',
  'concurrency',
  'Механизмы синхронизации в Go',
  'Какие из перечисленных примитивов синхронизации и подходов позволяют предотвратить состояние гонки (data race) при конкурентном доступе к разделяемой переменной?',
  '{}'::jsonb,
  '{
    "options": [
      {"id": "opt_mutex", "text": "Использование sync.Mutex / sync.RWMutex"},
      {"id": "opt_atomic", "text": "Атомарные операции из пакета sync/atomic"},
      {"id": "opt_channel", "text": "Передача владения данными через небуферизированный/буферизированный канал"},
      {"id": "opt_naked", "text": "Обычное чтение и запись переменной из нескольких горутин без блокировок"}
    ],
    "correct": ["opt_mutex", "opt_atomic", "opt_channel"]
  }'::jsonb,
  3, 1.3, 0.0
) ON CONFLICT (id) DO NOTHING;

-- 5. Middle: SQL-запрос с агрегацией и группировкой (SQL)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000005',
  '00000000-0000-0000-0000-000000000000',
  'sql',
  'databases',
  'Агрегация зарплат по отделам',
  'Напишите SQL-запрос для таблицы `employees` (`id`, `department_id`, `salary`).
Требуется получить `department_id` и среднюю зарплату (`avg_salary`) для отделов, в которых средняя зарплата превышает 150000. Результат отсортируйте по убыванию средней зарплаты.',
  '{}'::jsonb,
  '{
    "required_keywords": ["SELECT", "FROM", "GROUP BY", "HAVING", "AVG", "ORDER BY"],
    "target_tables": ["employees"]
  }'::jsonb,
  3, 1.4, 0.2
) ON CONFLICT (id) DO NOTHING;

-- 6. Middle: Регулярное выражение (Regex)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000006',
  '00000000-0000-0000-0000-000000000000',
  'regex',
  'regex',
  'Валидация семантической версии SemVer',
  'Напишите регулярное выражение для проверки строки семантической версии вида `MAJOR.MINOR.PATCH`, где каждая часть состоит из одной или более цифр (например, `1.0.4` или `12.3.456`). Начало и конец строки должны строго совпадать.',
  '{}'::jsonb,
  '{
    "matches": ["1.0.0", "0.12.3", "10.200.300"],
    "non_matches": ["1.0", "v1.0.0", "1.a.2", "1.0.0-beta"]
  }'::jsonb,
  3, 1.2, 0.3
) ON CONFLICT (id) DO NOTHING;

-- 7. Middle: Параметризованный вычислительный тест с динамическими аргументами (Античит)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000007',
  '00000000-0000-0000-0000-000000000000',
  'text',
  'algorithms',
  'Кольцевой буфер и индекс элемента',
  'В кольцевом буфере емкостью `N = {{capacity}}` головной указатель (head) равен `{{head}}`.
Требуется записать смещение `k = {{offset}}` элементов вперед по формуле `(head + k) % N`.
Каков результирующий индекс элемента в буфере?',
  '{
    "variables": {
      "capacity": {"type": "choice", "choices": ["8", "16", "32", "64"]},
      "head": {"type": "int_range", "min": 3, "max": 7},
      "offset": {"type": "int_range", "min": 10, "max": 25}
    },
    "formula": "head + offset"
  }'::jsonb,
  '{"expected": ""}'::jsonb,
  3, 1.1, 0.1
) ON CONFLICT (id) DO NOTHING;

-- 8. Senior: Паттерны отказоустойчивости (Single Choice)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000008',
  '00000000-0000-0000-0000-000000000000',
  'single_choice',
  'system_design',
  'Паттерн Circuit Breaker в распределенных системах',
  'В каком состоянии находится паттерн Circuit Breaker, когда он периодически пропускает ограниченное количество тестовых запросов к сбойному сервису для проверки восстановления его работоспособности?',
  '{}'::jsonb,
  '{
    "options": [
      {"id": "opt_half", "text": "Half-Open"},
      {"id": "opt_open", "text": "Open"},
      {"id": "opt_closed", "text": "Closed"},
      {"id": "opt_throttled", "text": "Throttled"}
    ],
    "correct": ["opt_half"]
  }'::jsonb,
  4, 1.5, 0.8
) ON CONFLICT (id) DO NOTHING;

-- 9. Senior: Уровни изоляции транзакций и аномалии (Single Choice)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000009',
  '00000000-0000-0000-0000-000000000000',
  'single_choice',
  'databases',
  'Аномалия фантомного чтения (Phantom Read)',
  'Какой минимальный уровень изоляции транзакций по стандарту ANSI/ISO SQL гарантирует предотвращение аномалии фантомного чтения (Phantom Read)?',
  '{}'::jsonb,
  '{
    "options": [
      {"id": "opt_serializable", "text": "Serializable"},
      {"id": "opt_repeatable", "text": "Repeatable Read"},
      {"id": "opt_read_committed", "text": "Read Committed"},
      {"id": "opt_read_uncommitted", "text": "Read Uncommitted"}
    ],
    "correct": ["opt_serializable"]
  }'::jsonb,
  4, 1.4, 0.9
) ON CONFLICT (id) DO NOTHING;

-- 10. Senior: Задача на безопасный код (Code)
INSERT INTO tasks (id, template_id, type, topic, title, body, generator, solution, difficulty, discrimination, difficulty_irt)
VALUES (
  'a1111111-0000-0000-0000-000000000010',
  '00000000-0000-0000-0000-000000000000',
  'code',
  'concurrency',
  'Потокобезопасный счетчик на Go',
  'Напишите структуру `SafeCounter` и метод `Inc()`, который увеличивает внутреннее целочисленное значение на 1 потокобезопасно с использованием `sync.Mutex`.',
  '{}'::jsonb,
  '{
    "required_patterns": ["sync.Mutex", "Lock()", "Unlock()"],
    "forbidden_tokens": ["panic"]
  }'::jsonb,
  5, 1.6, 1.4
) ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM tasks WHERE id LIKE 'a1111111-%';
