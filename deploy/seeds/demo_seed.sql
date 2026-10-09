-- ============================================================================
-- FSP Platform Demo Seed Data (BE2-05)
-- 80 Realistic Candidates with tests & FSP, 18 Employers, 26 Vacancies
-- ============================================================================

-- 1. Search read-model (candidate_search_docs)
TRUNCATE candidate_search_docs CASCADE;

INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001000', 'Иванова Алиса', '00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, TRUE, 'Без разряда',
  1644, 40.7, 1, 12, 12,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.8, 'г. Москва', 1,
  85.0, 53.7, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, Без разряда.', '{"fsp_verified"}'::text[], '2026-09-24T17:50:40.971224', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001001', 'Виноградов Александр', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.8, 'г. Екатеринбург', 1,
  85.0, 59.0, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-10T17:50:40.971266', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001002', 'Смирнов Кирилл', '00000000-0000-0000-0000-000000000505', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, TRUE, '2-й спортивный разряд',
  1794, 55.1, 2, 5, 11,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 1.8, 'г. Новосибирск', 0,
  85.0, 56.8, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, 2-й спортивный разряд.', '{"fsp_verified"}'::text[], '2026-08-23T17:50:40.971287', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001003', 'Виноградов Егор', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 2.2, 'г. Санкт-Петербург', 2,
  85.0, 60.9, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-16T17:50:40.971299', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001004', 'Петрова Алиса', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, TRUE, '2-й спортивный разряд',
  1719, 51.3, 2, 5, 12,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 1.6, 'г. Нижний Новгород', 2,
  85.0, 60.0, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, 2-й спортивный разряд.', '{"fsp_verified"}'::text[], '2026-09-28T17:50:40.971313', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001005', 'Соловьев Владимир', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.7, 'г. Екатеринбург', 0,
  85.0, 57.2, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-31T17:50:40.971322', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001006', 'Виноградов Алексей', '00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, TRUE, 'Без разряда',
  1712, 45.7, 1, 12, 11,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.1, 'г. Воронеж', 0,
  85.0, 54.0, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, Без разряда.', '{"fsp_verified"}'::text[], '2026-08-18T17:50:40.971334', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001007', 'Васильев Артем', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 0.9, 'г. Казань', 2,
  85.0, 63.9, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-25T17:50:40.971342', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001008', 'Новикова Алиса', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 2.0, 'г. Самара', 2,
  85.0, 60.9, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-22T17:50:40.971351', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001009', 'Богданов Роман', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, TRUE, '2-й спортивный разряд',
  1624, 44.0, 3, 5, 8,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 1.3, 'г. Самара', 3,
  85.0, 57.0, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, 2-й спортивный разряд.', '{"fsp_verified","active_solver"}'::text[], '2026-08-31T17:50:40.971361', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001010', 'Васильев Максим', '00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.6, 'г. Самара', 0,
  85.0, 57.2, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-26T17:50:40.971368', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001011', 'Виноградов Сергей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, TRUE, '3-й спортивный разряд',
  1680, 61.9, 3, 8, 6,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.9, 'г. Казань', 2,
  85.0, 61.2, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, 3-й спортивный разряд.', '{"fsp_verified"}'::text[], '2026-09-06T17:50:40.971378', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001012', 'Кузнецова Юлия', '00000000-0000-0000-0000-000000000505', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 2.1, 'г. Ростов-на-Дону', 1,
  85.0, 62.0, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-29T17:50:40.971386', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001013', 'Лебедев Владимир', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, TRUE, '2-й спортивный разряд',
  1906, 42.8, 1, 8, 13,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 2.1, 'г. Ростов-на-Дону', 2,
  85.0, 55.5, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, 2-й спортивный разряд.', '{"fsp_verified"}'::text[], '2026-08-13T17:50:40.971395', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001014', 'Новиков Никита', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 1.1, 'г. Москва', 0,
  85.0, 57.2, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-21T17:50:40.971403', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001015', 'Соколов Андрей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, TRUE, 'Без разряда',
  1910, 58.9, 3, 8, 9,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.5, 'г. Пермь', 1,
  85.0, 59.1, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, Без разряда.', '{"fsp_verified"}'::text[], '2026-08-24T17:50:40.971412', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001016', 'Козлова Екатерина', '00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 1.7, 'г. Пермь', 0,
  85.0, 57.2, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-17T17:50:40.971420', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001017', 'Богданов Александр', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 1.1, 'г. Ростов-на-Дону', 0,
  85.0, 57.2, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-24T17:50:40.971427', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001018', 'Иванов Максим', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000301', 'Junior', 2, 45.0, TRUE, 'Без разряда',
  1892, 59.6, 1, 12, 13,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 0.9, 'г. Нижний Новгород', 3,
  85.0, 63.7, 'Подтвержденный балл тестирования 45.0%, участник студенческой лиги ФСП, Без разряда.', '{"fsp_verified","active_solver"}'::text[], '2026-09-26T17:50:40.971437', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001019', 'Попов Сергей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000302', 'Junior+', 3, 49.4, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 1.4, 'г. Пермь', 0,
  85.0, 59.2, 'Подтвержденный балл тестирования 49.4%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-26T17:50:40.971444', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001020', 'Волкова Алиса', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 60.5, TRUE, 'КМС',
  2047, 71.1, 4, 5, 16,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 2.6, 'г. Санкт-Петербург', 3,
  85.0, 70.6, 'Подтвержденный балл тестирования 60.5%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_ranked","active_solver"}'::text[], '2026-09-23T17:50:40.971455', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001021', 'Павлов Максим', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 2.7, 'г. Ростов-на-Дону', 0,
  85.0, 60.2, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-10-03T17:50:40.971464', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001022', 'Лебедев Михаил', '00000000-0000-0000-0000-000000000506', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 66.0, TRUE, 'КМС',
  1951, 72.8, 4, 5, 18,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 3.6, 'г. Самара', 3,
  85.0, 75.0, 'Подтвержденный балл тестирования 66.0%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_ranked","active_solver"}'::text[], '2026-09-29T17:50:40.971474', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001023', 'Петров Иван', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 76.6, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 4.7, 'г. Екатеринбург', 1,
  85.0, 73.3, 'Подтвержденный балл тестирования 76.6%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-22T17:50:40.971482', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001024', 'Смирнова Дарья', '00000000-0000-0000-0000-000000000510', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 45.0, TRUE, 'КМС',
  2210, 78.1, 3, 2, 16,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 3.8, 'г. Москва', 1,
  85.0, 64.9, 'Подтвержденный балл тестирования 45.0%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_medalist","fsp_ranked"}'::text[], '2026-09-13T17:50:40.971495', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001025', 'Богданов Сергей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 64.6, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 3.8, 'г. Екатеринбург', 3,
  85.0, 71.5, 'Подтвержденный балл тестирования 64.6%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate","active_solver"}'::text[], '2026-08-27T17:50:40.971505', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001026', 'Морозов Артем', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 61.4, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 4.0, 'г. Екатеринбург', 2,
  85.0, 68.2, 'Подтвержденный балл тестирования 61.4%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-09T17:50:40.971513', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001027', 'Кузнецов Артем', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 74.9, TRUE, '1-й спортивный разряд',
  1987, 70.3, 4, 3, 22,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 3.6, 'г. Москва', 2,
  85.0, 74.2, 'Подтвержденный балл тестирования 74.9%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"fsp_verified","fsp_medalist","fsp_ranked"}'::text[], '2026-08-13T17:50:40.971524', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001028', 'Павлова Елена', '00000000-0000-0000-0000-000000000508', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 75.3, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 3.3, 'г. Екатеринбург', 7,
  85.0, 82.9, 'Подтвержденный балл тестирования 75.3%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate","active_solver"}'::text[], '2026-08-28T17:50:40.971533', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001029', 'Виноградов Александр', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 82.8, TRUE, '1-й спортивный разряд',
  2018, 80.6, 3, 4, 15,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 4.6, 'г. Казань', 2,
  85.0, 80.1, 'Подтвержденный балл тестирования 82.8%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"high_test_score","fsp_verified","fsp_ranked"}'::text[], '2026-09-21T17:50:40.971542', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001030', 'Воробьев Никита', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 69.4, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 4.2, 'г. Екатеринбург', 2,
  85.0, 71.8, 'Подтвержденный балл тестирования 69.4%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-12T17:50:40.971550', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001031', 'Кузнецов Дмитрий', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 77.3, TRUE, '1-й спортивный разряд',
  2120, 70.7, 4, 3, 24,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 3.5, 'г. Пермь', 4,
  85.0, 77.6, 'Подтвержденный балл тестирования 77.3%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"fsp_verified","fsp_medalist","fsp_ranked","active_solver"}'::text[], '2026-09-03T17:50:40.971559', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001032', 'Козлова София', '00000000-0000-0000-0000-000000000506', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 83.3, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 2.5, 'г. Ростов-на-Дону', 7,
  85.0, 86.5, 'Подтвержденный балл тестирования 83.3%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate","active_solver"}'::text[], '2026-09-15T17:50:40.971569', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001033', 'Зайцев Андрей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 76.8, TRUE, '1-й спортивный разряд',
  1970, 72.2, 5, 4, 17,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 2.6, 'г. Новосибирск', 1,
  85.0, 74.2, 'Подтвержденный балл тестирования 76.8%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-09-16T17:50:40.971579', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001034', 'Воробьев Роман', '00000000-0000-0000-0000-000000000510', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 84.4, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 2.9, 'г. Воронеж', 1,
  85.0, 79.8, 'Подтвержденный балл тестирования 84.4%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate"}'::text[], '2026-09-27T17:50:40.971587', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001035', 'Зайцев Артем', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 52.5, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 4.5, 'г. Пермь', 0,
  85.0, 60.6, 'Подтвержденный балл тестирования 52.5%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-14T17:50:40.971594', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001036', 'Зайцева Анна', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 71.2, TRUE, 'КМС',
  2066, 71.3, 5, 3, 14,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 3.0, 'г. Томск', 2,
  85.0, 73.2, 'Подтвержденный балл тестирования 71.2%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_medalist","fsp_ranked"}'::text[], '2026-09-21T17:50:40.971604', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001037', 'Морозов Максим', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 52.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 4.0, 'г. Нижний Новгород', 2,
  85.0, 67.0, 'Подтвержденный балл тестирования 52.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-10-07T17:50:40.971611', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001038', 'Морозов Сергей', '00000000-0000-0000-0000-000000000508', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 67.7, TRUE, 'КМС',
  2323, 84.7, 5, 4, 20,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 3.8, 'г. Новосибирск', 0,
  85.0, 73.6, 'Подтвержденный балл тестирования 67.7%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-09-14T17:50:40.971621', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001039', 'Морозов Иван', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 64.3, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 4.1, 'г. Москва', 1,
  85.0, 67.7, 'Подтвержденный балл тестирования 64.3%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-15T17:50:40.971628', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001040', 'Иванова Анастасия', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 51.6, TRUE, 'КМС',
  2103, 73.5, 4, 4, 24,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 4.6, 'г. Пермь', 2,
  85.0, 67.0, 'Подтвержденный балл тестирования 51.6%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-09-03T17:50:40.971637', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001041', 'Козлов Андрей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 51.8, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 4.0, 'г. Самара', 1,
  85.0, 62.1, 'Подтвержденный балл тестирования 51.8%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-30T17:50:40.971645', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001042', 'Васильев Никита', '00000000-0000-0000-0000-000000000506', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 62.0, TRUE, 'КМС',
  2285, 75.5, 4, 4, 21,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 4.4, 'г. Ростов-на-Дону', 1,
  85.0, 70.0, 'Подтвержденный балл тестирования 62.0%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-09-06T17:50:40.971654', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001043', 'Лебедев Илья', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 72.8, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 2.7, 'г. Пермь', 2,
  85.0, 76.4, 'Подтвержденный балл тестирования 72.8%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-10-03T17:50:40.971662', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001044', 'Попова Юлия', '00000000-0000-0000-0000-000000000510', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 65.9, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 3.2, 'г. Пермь', 3,
  85.0, 72.1, 'Подтвержденный балл тестирования 65.9%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate","active_solver"}'::text[], '2026-08-30T17:50:40.971669', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001045', 'Павлов Максим', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 56.6, TRUE, '1-й спортивный разряд',
  2306, 72.1, 4, 5, 21,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 4.5, 'г. Самара', 0,
  85.0, 65.9, 'Подтвержденный балл тестирования 56.6%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-08-12T17:50:40.971679', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001046', 'Зайцев Сергей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 77.6, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 2.9, 'г. Екатеринбург', 7,
  85.0, 83.9, 'Подтвержденный балл тестирования 77.6%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate","active_solver"}'::text[], '2026-10-01T17:50:40.971687', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001047', 'Соколов Егор', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 61.0, TRUE, '1-й спортивный разряд',
  2336, 83.2, 5, 5, 23,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 4.0, 'г. Томск', 3,
  85.0, 74.4, 'Подтвержденный балл тестирования 61.0%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"fsp_verified","fsp_ranked","active_solver"}'::text[], '2026-08-16T17:50:40.971696', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001048', 'Новикова Ксения', '00000000-0000-0000-0000-000000000508', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 56.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 4.2, 'г. Новосибирск', 2,
  85.0, 65.8, 'Подтвержденный балл тестирования 56.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-20T17:50:40.971704', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001049', 'Новиков Илья', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 65.2, TRUE, '1-й спортивный разряд',
  2089, 69.4, 5, 4, 19,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 3.5, 'г. Казань', 1,
  85.0, 69.3, 'Подтвержденный балл тестирования 65.2%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-09-24T17:50:40.971712', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001050', 'Соколов Павел', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 54.2, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 3.0, 'г. Воронеж', 1,
  85.0, 63.2, 'Подтвержденный балл тестирования 54.2%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-16T17:50:40.971720', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001051', 'Васильев Роман', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 71.4, TRUE, '1-й спортивный разряд',
  2341, 84.0, 3, 5, 21,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 4.7, 'г. Уфа', 2,
  85.0, 77.1, 'Подтвержденный балл тестирования 71.4%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-08-21T17:50:40.971728', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001052', 'Зайцева Анастасия', '00000000-0000-0000-0000-000000000506', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 62.4, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 3.7, 'г. Самара', 2,
  85.0, 68.7, 'Подтвержденный балл тестирования 62.4%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-09-11T17:50:40.971736', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001053', 'Смирнов Илья', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000304', 'Middle+', 5, 45.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 3.3, 'г. Самара', 1,
  85.0, 59.0, 'Подтвержденный балл тестирования 45.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate"}'::text[], '2026-08-16T17:50:40.971743', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001054', 'Соколов Егор', '00000000-0000-0000-0000-000000000510', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000303', 'Middle', 4, 75.0, TRUE, 'КМС',
  2279, 71.1, 4, 5, 16,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 3.7, 'г. Уфа', 2,
  85.0, 74.5, 'Подтвержденный балл тестирования 75.0%, призёр Всероссийского хакатона ФСП, КМС.', '{"fsp_verified","fsp_ranked"}'::text[], '2026-09-14T17:50:40.971752', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001055', 'Козлов Артем', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 81.8, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 6.0, 'г. Томск', 7,
  85.0, 85.8, 'Подтвержденный балл тестирования 81.8%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate","active_solver"}'::text[], '2026-09-12T17:50:40.971760', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001056', 'Зайцева Екатерина', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 86.1, TRUE, 'КМС',
  1985, 71.4, 3, 2, 14,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 6.4, 'г. Санкт-Петербург', 1,
  85.0, 77.3, 'Подтвержденный балл тестирования 86.1%, призёр Всероссийского хакатона ФСП, КМС.', '{"high_test_score","fsp_verified","fsp_medalist","fsp_ranked"}'::text[], '2026-08-30T17:50:40.971770', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001057', 'Новиков Андрей', '00000000-0000-0000-0000-000000000507', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 85.6, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 6.4, 'г. Новосибирск', 2,
  85.0, 79.1, 'Подтвержденный балл тестирования 85.6%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate"}'::text[], '2026-09-09T17:50:40.971778', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001058', 'Соловьев Кирилл', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 95.3, TRUE, 'КМС',
  2005, 79.5, 5, 2, 18,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 6.8, 'г. Новосибирск', 4,
  85.0, 86.5, 'Топ-5% по тесту Mobile Developer (95.3%), призёр Всероссийского хакатона ФСП, КМС.', '{"top_test_performer","fsp_verified","fsp_medalist","fsp_ranked","active_solver"}'::text[], '2026-09-13T17:50:40.971787', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001059', 'Кузнецов Иван', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 84.1, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 7.1, 'г. Уфа', 6,
  85.0, 85.6, 'Подтвержденный балл тестирования 84.1%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate","active_solver"}'::text[], '2026-08-20T17:50:40.971795', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001060', 'Зайцева Екатерина', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 94.3, TRUE, 'Мастер спорта',
  2459, 85.8, 6, 1, 34,
  '{"победитель Чемпионата России ФСП 2024"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 6.8, 'г. Пермь', 6,
  85.0, 90.4, 'Топ-5% по тесту Backend Go (94.3%), победитель Чемпионата России ФСП 2024, Мастер спорта.', '{"top_test_performer","fsp_verified","fsp_champion","fsp_master","active_solver"}'::text[], '2026-09-17T17:50:40.971805', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001061', 'Зайцев Александр', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 95.2, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 5.3, 'г. Томск', 7,
  85.0, 91.8, 'Топ-5% по тесту Backend Python (95.2%), без истории соревнований ФСП (оценка по тестам и стеку).', '{"top_test_performer","no_fsp_candidate","active_solver"}'::text[], '2026-09-09T17:50:40.971813', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001062', 'Зайцев Андрей', '00000000-0000-0000-0000-000000000507', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 91.9, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 7.2, 'г. Новосибирск', 4,
  85.0, 85.6, 'Топ-5% по тесту Frontend React (91.9%), без истории соревнований ФСП (оценка по тестам и стеку).', '{"top_test_performer","no_fsp_candidate","active_solver"}'::text[], '2026-09-11T17:50:40.971821', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001063', 'Волков Кирилл', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 85.4, TRUE, 'Мастер спорта',
  2465, 91.3, 7, 2, 27,
  '{"серебряный призёр Кубка ФСП 2025"}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 7.5, 'г. Екатеринбург', 6,
  85.0, 89.0, 'Подтвержденный балл тестирования 85.4%, серебряный призёр Кубка ФСП 2025, Мастер спорта.', '{"high_test_score","fsp_verified","fsp_medalist","fsp_master","active_solver"}'::text[], '2026-09-14T17:50:40.971830', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001064', 'Смирнова Дарья', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 88.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 7.6, 'г. Томск', 5,
  85.0, 85.6, 'Подтвержденный балл тестирования 88.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate","active_solver"}'::text[], '2026-08-25T17:50:40.971838', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001065', 'Виноградов Кирилл', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 88.6, TRUE, '1-й спортивный разряд',
  2318, 80.9, 5, 5, 21,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 6.5, 'г. Москва', 6,
  85.0, 87.0, 'Подтвержденный балл тестирования 88.6%, призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"high_test_score","fsp_verified","fsp_ranked","active_solver"}'::text[], '2026-09-08T17:50:40.971847', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001066', 'Павлов Илья', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 89.3, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 5.3, 'г. Москва', 3,
  85.0, 82.6, 'Подтвержденный балл тестирования 89.3%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate","active_solver"}'::text[], '2026-09-08T17:50:40.971854', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001067', 'Зайцев Владимир', '00000000-0000-0000-0000-000000000507', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 94.2, TRUE, '1-й спортивный разряд',
  2088, 70.1, 3, 4, 18,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 6.7, 'г. Воронеж', 2,
  85.0, 80.9, 'Топ-5% по тесту Frontend Vue (94.2%), призёр Всероссийского хакатона ФСП, 1-й спортивный разряд.', '{"top_test_performer","fsp_verified","fsp_ranked"}'::text[], '2026-09-18T17:50:40.971864', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001068', 'Соловьева Елена', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 85.0, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 7.9, 'г. Ростов-на-Дону', 5,
  85.0, 84.2, 'Подтвержденный балл тестирования 85.0%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate","active_solver"}'::text[], '2026-08-21T17:50:40.971873', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001069', 'Попов Никита', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 82.3, TRUE, 'Мастер спорта',
  2353, 88.0, 5, 2, 35,
  '{"серебряный призёр Кубка ФСП 2025"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 5.9, 'г. Екатеринбург', 1,
  85.0, 80.9, 'Подтвержденный балл тестирования 82.3%, серебряный призёр Кубка ФСП 2025, Мастер спорта.', '{"high_test_score","fsp_verified","fsp_medalist","fsp_master"}'::text[], '2026-09-03T17:50:40.971882', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001070', 'Соколов Никита', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 71.2, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 7.6, 'г. Пермь', 3,
  85.0, 74.4, 'Подтвержденный балл тестирования 71.2%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"no_fsp_candidate","active_solver"}'::text[], '2026-09-10T17:50:40.971889', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001071', 'Лебедев Артем', '00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 86.3, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 5.8, 'г. Москва', 1,
  85.0, 77.6, 'Подтвержденный балл тестирования 86.3%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate"}'::text[], '2026-09-13T17:50:40.971896', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001072', 'Иванова Полина', '00000000-0000-0000-0000-000000000507', '00000000-0000-0000-0000-000000000202', 'Frontend React',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 91.6, TRUE, 'Мастер спорта',
  2371, 91.1, 4, 1, 27,
  '{"победитель Чемпионата России ФСП 2024"}'::text[], '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406,00000000-0000-0000-0000-000000000407}'::uuid[], 6.9, 'г. Уфа', 4,
  85.0, 88.7, 'Топ-5% по тесту Frontend React (91.6%), победитель Чемпионата России ФСП 2024, Мастер спорта.', '{"top_test_performer","fsp_verified","fsp_champion","fsp_master","active_solver"}'::text[], '2026-08-31T17:50:40.971905', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001073', 'Голубев Михаил', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 87.4, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000411}'::uuid[], 6.4, 'г. Самара', 5,
  85.0, 85.3, 'Подтвержденный балл тестирования 87.4%, без истории соревнований ФСП (оценка по тестам и стеку).', '{"high_test_score","no_fsp_candidate","active_solver"}'::text[], '2026-09-11T17:50:40.971913', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001074', 'Богданов Никита', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000305', 'Senior', 6, 89.9, TRUE, 'КМС',
  2085, 71.0, 3, 2, 16,
  '{"призёр Всероссийского хакатона ФСП"}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 5.2, 'г. Уфа', 1,
  85.0, 80.5, 'Подтвержденный балл тестирования 89.9%, призёр Всероссийского хакатона ФСП, КМС.', '{"high_test_score","fsp_verified","fsp_medalist","fsp_ranked"}'::text[], '2026-09-28T17:50:40.971924', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001075', 'Зайцев Александр', '00000000-0000-0000-0000-000000000504', '00000000-0000-0000-0000-000000000201', 'Backend Python',
  '00000000-0000-0000-0000-000000000306', 'Lead', 7, 91.4, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 10.8, 'г. Томск', 2,
  85.0, 81.7, 'Топ-5% по тесту Backend Python (91.4%), без истории соревнований ФСП (оценка по тестам и стеку).', '{"top_test_performer","no_fsp_candidate"}'::text[], '2026-09-20T17:50:40.971932', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001076', 'Соколова София', '00000000-0000-0000-0000-000000000504', '00000000-0000-0000-0000-000000000201', 'Backend Go',
  '00000000-0000-0000-0000-000000000306', 'Lead', 7, 95.4, TRUE, 'Без разряда',
  1901, 42.9, 2, 12, 9,
  '{"участник студенческой лиги ФСП"}'::text[], '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403,00000000-0000-0000-0000-000000000407}'::uuid[], 11.4, 'г. Воронеж', 2,
  85.0, 73.2, 'Топ-5% по тесту Backend Go (95.4%), участник студенческой лиги ФСП, Без разряда.', '{"top_test_performer","fsp_verified"}'::text[], '2026-08-28T17:50:40.971941', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001077', 'Морозов Андрей', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000202', 'Frontend Vue',
  '00000000-0000-0000-0000-000000000306', 'Lead', 7, 94.6, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000409,00000000-0000-0000-0000-000000000406}'::uuid[], 8.3, 'г. Новосибирск', 5,
  85.0, 88.6, 'Топ-5% по тесту Frontend Vue (94.6%), без истории соревнований ФСП (оценка по тестам и стеку).', '{"top_test_performer","no_fsp_candidate","active_solver"}'::text[], '2026-08-22T17:50:40.971951', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001078', 'Павлов Никита', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000203', 'Mobile Developer',
  '00000000-0000-0000-0000-000000000306', 'Lead', 7, 91.5, TRUE, 'Мастер спорта',
  2476, 91.6, 4, 1, 31,
  '{"победитель Чемпионата России ФСП 2024"}'::text[], '{00000000-0000-0000-0000-000000000412}'::uuid[], 9.9, 'г. Санкт-Петербург', 4,
  85.0, 88.8, 'Топ-5% по тесту Mobile Developer (91.5%), победитель Чемпионата России ФСП 2024, Мастер спорта.', '{"top_test_performer","fsp_verified","fsp_champion","fsp_master","active_solver"}'::text[], '2026-08-22T17:50:40.971960', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;
INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '22222222-2222-2222-2222-000000001079', 'Воробьев Артем', '00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000205', 'Data / AI Developer',
  '00000000-0000-0000-0000-000000000306', 'Lead', 7, 96.7, FALSE, NULL,
  0, 0.0, 0, NULL, 0,
  '{}'::text[], '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 8.1, 'г. Казань', 7,
  85.0, 92.5, 'Топ-5% по тесту Data / AI Developer (96.7%), без истории соревнований ФСП (оценка по тестам и стеку).', '{"top_test_performer","no_fsp_candidate","active_solver"}'::text[], '2026-08-14T17:50:40.971968', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;

-- 2. Companies & Vacancies
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001000', '44444444-4444-4444-4444-000000001000', 'Яндекс Финтех', 'Разработка высоконагруженных платежных сервисов и банковских продуктов.', 'Финтех / Платежи', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001000', '33333333-3333-3333-3333-000000001000', 'Senior Go Developer', 'Описание вакансии Senior Go Developer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000305', '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 320000, 480000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001001', '44444444-4444-4444-4444-000000001001', 'Т-Банк (Тинькофф)', 'Экосистема финансовых и лайфстайл сервисов для миллионов пользователей.', 'Банки / Финтех', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001001', '33333333-3333-3333-3333-000000001001', 'Middle Python Backend Engineer', 'Описание вакансии Middle Python Backend Engineer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000407}'::uuid[], 200000, 280000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001002', '44444444-4444-4444-4444-000000001002', 'СберТех', 'Инновационная облачная платформа цифровизации для банковского сектора.', 'Корпоративные платформы', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001002', '33333333-3333-3333-3333-000000001002', 'Frontend React / TypeScript Developer', 'Описание вакансии Frontend React / TypeScript Developer', '00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406}'::uuid[], 190000, 260000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001003', '44444444-4444-4444-4444-000000001003', 'Ozon E-commerce', 'Маркетплейс номер один: миллионы RPS, распределенные склады и микросервисы.', 'E-commerce', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001003', '33333333-3333-3333-3333-000000001003', 'Senior Go Developer', 'Описание вакансии Senior Go Developer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000305', '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 320000, 480000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001004', '44444444-4444-4444-4444-000000001004', 'Авито Платформа', 'Самый посещаемый классифайд в мире с передовым технологическим стеком.', 'Классифайд', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001004', '33333333-3333-3333-3333-000000001004', 'Middle Python Backend Engineer', 'Описание вакансии Middle Python Backend Engineer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000407}'::uuid[], 200000, 280000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001005', '44444444-4444-4444-4444-000000001005', 'VK Tech', 'Корпоративные облачные сервисы, коммуникации и медиаплатформы.', 'Социальные сети / Cloud', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001005', '33333333-3333-3333-3333-000000001005', 'Frontend React / TypeScript Developer', 'Описание вакансии Frontend React / TypeScript Developer', '00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406}'::uuid[], 190000, 260000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001006', '44444444-4444-4444-4444-000000001006', 'Лаборатория Касперского', 'Мировой лидер в области кибербезопасности и защиты данных.', 'Кибербезопасность', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001006', '33333333-3333-3333-3333-000000001006', 'Senior Go Developer', 'Описание вакансии Senior Go Developer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000305', '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 320000, 480000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001007', '44444444-4444-4444-4444-000000001007', '2ГИС Геосервисы', 'Детализированные 3D-карты и навигация для городов России и мира.', 'Карты / Навигация', '500-1000')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001007', '33333333-3333-3333-3333-000000001007', 'Middle Python Backend Engineer', 'Описание вакансии Middle Python Backend Engineer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000407}'::uuid[], 200000, 280000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001008', '44444444-4444-4444-4444-000000001008', 'Selectel', 'Ведущий провайдер IT-инфраструктуры и облачных платформ в РФ.', 'Облачные сервисы / IaaS', '500-1000')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001008', '33333333-3333-3333-3333-000000001008', 'Frontend React / TypeScript Developer', 'Описание вакансии Frontend React / TypeScript Developer', '00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406}'::uuid[], 190000, 260000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001009', '44444444-4444-4444-4444-000000001009', 'Positive Technologies', 'Продукты результативной кибербезопасности и обнаружения угроз.', 'Информационная безопасность', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001009', '33333333-3333-3333-3333-000000001009', 'Senior Go Developer', 'Описание вакансии Senior Go Developer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000305', '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 320000, 480000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001010', '44444444-4444-4444-4444-000000001010', 'Wildberries Инфраструктура', 'Высоконагруженная логистическая и торговая платформа.', 'E-commerce', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001010', '33333333-3333-3333-3333-000000001010', 'Middle Python Backend Engineer', 'Описание вакансии Middle Python Backend Engineer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000407}'::uuid[], 200000, 280000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001011', '44444444-4444-4444-4444-000000001011', 'Skyeng EdTech', 'Крупнейшая EdTech компания в Восточной Европе с ИИ-технологиями обучения.', 'Образование', '500-1000')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001011', '33333333-3333-3333-3333-000000001011', 'Frontend React / TypeScript Developer', 'Описание вакансии Frontend React / TypeScript Developer', '00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406}'::uuid[], 190000, 260000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001012', '44444444-4444-4444-4444-000000001012', 'МойОфис', 'Безопасные офисные решения для корпоративных коммуникаций.', 'Офисное ПО', '500-1000')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001012', '33333333-3333-3333-3333-000000001012', 'Senior Go Developer', 'Описание вакансии Senior Go Developer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000305', '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 320000, 480000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001013', '44444444-4444-4444-4444-000000001013', 'Киберпротек', 'Решения для резервного копирования и защиты данных от киберугроз.', 'Резервное копирование', '100-500')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001013', '33333333-3333-3333-3333-000000001013', 'Middle Python Backend Engineer', 'Описание вакансии Middle Python Backend Engineer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000407}'::uuid[], 200000, 280000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001014', '44444444-4444-4444-4444-000000001014', 'Купер (СберМаркет)', 'Сервис быстрой доставки продуктов и товаров первой необходимости.', 'FoodTech / Доставка', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001014', '33333333-3333-3333-3333-000000001014', 'Frontend React / TypeScript Developer', 'Описание вакансии Frontend React / TypeScript Developer', '00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406}'::uuid[], 190000, 260000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001015', '44444444-4444-4444-4444-000000001015', 'Yadro', 'Разработка аппаратных платформ, СХД и телеком-оборудования.', 'Высокопроизводительные серверы', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001015', '33333333-3333-3333-3333-000000001015', 'Senior Go Developer', 'Описание вакансии Senior Go Developer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000305', '{00000000-0000-0000-0000-000000000401,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000403}'::uuid[], 320000, 480000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001016', '44444444-4444-4444-4444-000000001016', 'Самокат', 'Сервис сверхбыстрой доставки за 15 минут в десятках городов.', 'E-grocery / Логистика', '1000+')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001016', '33333333-3333-3333-3333-000000001016', 'Middle Python Backend Engineer', 'Описание вакансии Middle Python Backend Engineer', '00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000404,00000000-0000-0000-0000-000000000402,00000000-0000-0000-0000-000000000407}'::uuid[], 200000, 280000, 'RUB', 'remote', 'published')
ON CONFLICT (id) DO NOTHING;
INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('33333333-3333-3333-3333-000000001017', '44444444-4444-4444-4444-000000001017', 'Mindbox', 'Автоматизация маркетинга и клиентских данных для enterprise-бизнеса.', 'Маркетинг / MarTech', '100-500')
ON CONFLICT (id) DO NOTHING;
INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('55555555-5555-5555-5555-000000001017', '33333333-3333-3333-3333-000000001017', 'Frontend React / TypeScript Developer', 'Описание вакансии Frontend React / TypeScript Developer', '00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000303', '{00000000-0000-0000-0000-000000000405,00000000-0000-0000-0000-000000000406}'::uuid[], 190000, 260000, 'RUB', 'hybrid', 'published')
ON CONFLICT (id) DO NOTHING;
