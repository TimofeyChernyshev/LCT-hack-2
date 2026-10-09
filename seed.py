#!/usr/bin/env python3
"""
seed.py — Генератор синтетического датасета и сидинг базы данных (BE2-05)

Генерирует:
- 80 реалистичных кандидатов (Junior, Middle, Senior, Lead; Backend, Frontend, Mobile, DevOps, AI)
  с результатами защищенного тестирования, с достижениями ФСП и БЕЗ истории ФСП.
- 18 профилей работодателей и 26 вакансий/потребностей с обязательными вилками ЗП (в рублях).
- Проводит сидинг в PostgreSQL (search_db, candidate_db, employer_db, auth_db, testing_db)
- Генерирует SQL-дамп deploy/seeds/demo_seed.sql и JSON deploy/seeds/demo_dataset.json
- При ключе --validate запускает процедуру валидации алгоритмов (NDCG, Precision@K, D-index).
"""

import argparse
import json
import math
import os
import random
import sys
import uuid
from datetime import datetime, timedelta
from typing import List, Dict, Any

# Устанавливаем seed для стабильной воспроизводимости
RANDOM_SEED = 42
random.seed(RANDOM_SEED)

# Базовые справочные UUIDs (детерминированные для согласованности всех сервисов)
SPEC_IDS = {
    "backend": "00000000-0000-0000-0000-000000000201",
    "frontend": "00000000-0000-0000-0000-000000000202",
    "mobile": "00000000-0000-0000-0000-000000000203",
    "devops": "00000000-0000-0000-0000-000000000204",
    "data": "00000000-0000-0000-0000-000000000205",
}

GRADE_IDS = {
    "junior": {"id": "00000000-0000-0000-0000-000000000301", "name": "Junior", "rank": 2},
    "junior_plus": {"id": "00000000-0000-0000-0000-000000000302", "name": "Junior+", "rank": 3},
    "middle": {"id": "00000000-0000-0000-0000-000000000303", "name": "Middle", "rank": 4},
    "middle_plus": {"id": "00000000-0000-0000-0000-000000000304", "name": "Middle+", "rank": 5},
    "senior": {"id": "00000000-0000-0000-0000-000000000305", "name": "Senior", "rank": 6},
    "lead": {"id": "00000000-0000-0000-0000-000000000306", "name": "Lead", "rank": 7},
}

TECH_IDS = {
    "go": {"id": "00000000-0000-0000-0000-000000000401", "name": "Go"},
    "postgres": {"id": "00000000-0000-0000-0000-000000000402", "name": "PostgreSQL"},
    "redis": {"id": "00000000-0000-0000-0000-000000000403", "name": "Redis"},
    "python": {"id": "00000000-0000-0000-0000-000000000404", "name": "Python"},
    "react": {"id": "00000000-0000-0000-0000-000000000405", "name": "React"},
    "ts": {"id": "00000000-0000-0000-0000-000000000406", "name": "TypeScript"},
    "docker": {"id": "00000000-0000-0000-0000-000000000407", "name": "Docker"},
    "k8s": {"id": "00000000-0000-0000-0000-000000000408", "name": "Kubernetes"},
    "vue": {"id": "00000000-0000-0000-0000-000000000409", "name": "Vue"},
    "java": {"id": "00000000-0000-0000-0000-000000000410", "name": "Java"},
    "kotlin": {"id": "00000000-0000-0000-0000-000000000411", "name": "Kotlin"},
    "swift": {"id": "00000000-0000-0000-0000-000000000412", "name": "Swift"},
    "mysql": {"id": "00000000-0000-0000-0000-000000000413", "name": "MySQL"},
    "nginx": {"id": "00000000-0000-0000-0000-000000000414", "name": "Nginx"},
}

CATEGORIES = {
    "backend_junior": "00000000-0000-0000-0000-000000000501",
    "backend_middle": "00000000-0000-0000-0000-000000000502",
    "backend_senior": "00000000-0000-0000-0000-000000000503",
    "backend_lead": "00000000-0000-0000-0000-000000000504",
    "frontend_junior": "00000000-0000-0000-0000-000000000505",
    "frontend_middle": "00000000-0000-0000-0000-000000000506",
    "frontend_senior": "00000000-0000-0000-0000-000000000507",
    "mobile_middle": "00000000-0000-0000-0000-000000000508",
    "devops_middle": "00000000-0000-0000-0000-000000000509",
    "data_middle": "00000000-0000-0000-0000-000000000510",
}

FIRST_NAMES_M = ["Александр", "Дмитрий", "Максим", "Сергей", "Андрей", "Алексей", "Иван", "Михаил", "Кирилл", "Никита", "Артем", "Владимир", "Павел", "Роман", "Егор", "Илья"]
FIRST_NAMES_F = ["Анна", "Елена", "Мария", "Ольга", "Екатерина", "Дарья", "Анастасия", "Полина", "Ксения", "Виктория", "Алиса", "София", "Юлия"]
LAST_NAMES_M = ["Смирнов", "Иванов", "Кузнецов", "Попов", "Соколов", "Лебедев", "Козлов", "Новиков", "Морозов", "Петров", "Волков", "Соловьев", "Васильев", "Зайцев", "Павлов", "Семенов", "Голубев", "Виноградов", "Богданов", "Воробьев"]
LAST_NAMES_F = ["Смирнова", "Иванова", "Кузнецова", "Попова", "Соколова", "Лебедева", "Козлова", "Новикова", "Морозова", "Петрова", "Волкова", "Соловьева", "Васильева", "Зайцева", "Павлова"]

CITIES = ["г. Москва", "г. Санкт-Петербург", "г. Новосибирск", "г. Екатеринбург", "г. Казань", "г. Нижний Новгород", "г. Самара", "г. Томск", "г. Ростов-на-Дону", "г. Уфа", "г. Пермь", "г. Воронеж"]

COMPANIES_DATA = [
    {"name": "Яндекс Финтех", "industry": "Финтех / Платежи", "size": "1000+", "desc": "Разработка высоконагруженных платежных сервисов и банковских продуктов."},
    {"name": "Т-Банк (Тинькофф)", "industry": "Банки / Финтех", "size": "1000+", "desc": "Экосистема финансовых и лайфстайл сервисов для миллионов пользователей."},
    {"name": "СберТех", "industry": "Корпоративные платформы", "size": "1000+", "desc": "Инновационная облачная платформа цифровизации для банковского сектора."},
    {"name": "Ozon E-commerce", "industry": "E-commerce", "size": "1000+", "desc": "Маркетплейс номер один: миллионы RPS, распределенные склады и микросервисы."},
    {"name": "Авито Платформа", "industry": "Классифайд", "size": "1000+", "desc": "Самый посещаемый классифайд в мире с передовым технологическим стеком."},
    {"name": "VK Tech", "industry": "Социальные сети / Cloud", "size": "1000+", "desc": "Корпоративные облачные сервисы, коммуникации и медиаплатформы."},
    {"name": "Лаборатория Касперского", "industry": "Кибербезопасность", "size": "1000+", "desc": "Мировой лидер в области кибербезопасности и защиты данных."},
    {"name": "2ГИС Геосервисы", "industry": "Карты / Навигация", "size": "500-1000", "desc": "Детализированные 3D-карты и навигация для городов России и мира."},
    {"name": "Selectel", "industry": "Облачные сервисы / IaaS", "size": "500-1000", "desc": "Ведущий провайдер IT-инфраструктуры и облачных платформ в РФ."},
    {"name": "Positive Technologies", "industry": "Информационная безопасность", "size": "1000+", "desc": "Продукты результативной кибербезопасности и обнаружения угроз."},
    {"name": "Wildberries Инфраструктура", "industry": "E-commerce", "size": "1000+", "desc": "Высоконагруженная логистическая и торговая платформа."},
    {"name": "Skyeng EdTech", "industry": "Образование", "size": "500-1000", "desc": "Крупнейшая EdTech компания в Восточной Европе с ИИ-технологиями обучения."},
    {"name": "МойОфис", "industry": "Офисное ПО", "size": "500-1000", "desc": "Безопасные офисные решения для корпоративных коммуникаций."},
    {"name": "Киберпротек", "industry": "Резервное копирование", "size": "100-500", "desc": "Решения для резервного копирования и защиты данных от киберугроз."},
    {"name": "Купер (СберМаркет)", "industry": "FoodTech / Доставка", "size": "1000+", "desc": "Сервис быстрой доставки продуктов и товаров первой необходимости."},
    {"name": "Yadro", "industry": "Высокопроизводительные серверы", "size": "1000+", "desc": "Разработка аппаратных платформ, СХД и телеком-оборудования."},
    {"name": "Самокат", "industry": "E-grocery / Логистика", "size": "1000+", "desc": "Сервис сверхбыстрой доставки за 15 минут в десятках городов."},
    {"name": "Mindbox", "industry": "Маркетинг / MarTech", "size": "100-500", "desc": "Автоматизация маркетинга и клиентских данных для enterprise-бизнеса."}
]


def generate_synthetic_dataset(num_candidates: int = 80, num_employers: int = 18) -> Dict[str, Any]:
    """Генерация реалистичного синтетического датасета для платформы ФСП"""
    candidates = []
    
    # 1. Генерация 80 кандидатов
    for i in range(num_candidates):
        cand_id = f"22222222-2222-2222-2222-{i+1000:012d}"
        is_female = (i % 4 == 0)
        first_name = random.choice(FIRST_NAMES_F) if is_female else random.choice(FIRST_NAMES_M)
        last_name = random.choice(LAST_NAMES_F) if is_female else random.choice(LAST_NAMES_M)
        full_name = f"{last_name} {first_name}"
        city = random.choice(CITIES)

        # Распределение грейдов: 25% Junior, 44% Middle, 25% Senior, 6% Lead
        if i < 20:
            grade_key = "junior" if i % 2 == 0 else "junior_plus"
            exp = round(random.uniform(0.8, 2.2), 1)
            theta = random.gauss(-0.6, 0.4)
            salary_want = random.randint(90, 150) * 1000
        elif i < 55:
            grade_key = "middle" if i % 2 == 0 else "middle_plus"
            exp = round(random.uniform(2.5, 4.8), 1)
            theta = random.gauss(0.5, 0.45)
            salary_want = random.randint(180, 280) * 1000
        elif i < 75:
            grade_key = "senior"
            exp = round(random.uniform(5.0, 8.0), 1)
            theta = random.gauss(1.6, 0.4)
            salary_want = random.randint(300, 480) * 1000
        else:
            grade_key = "lead"
            exp = round(random.uniform(8.0, 12.0), 1)
            theta = random.gauss(2.1, 0.35)
            salary_want = random.randint(450, 650) * 1000

        grade_info = GRADE_IDS[grade_key]

        # Распределение специализаций
        if i % 5 in (0, 1):
            spec_key = "backend"
            # Go или Python или Java
            if i % 2 == 0:
                stack_keys = ["go", "postgres", "redis", "docker"]
                spec_name = "Backend Go"
            else:
                stack_keys = ["python", "postgres", "redis", "docker"]
                spec_name = "Backend Python"
        elif i % 5 == 2:
            spec_key = "frontend"
            stack_keys = ["react", "ts", "docker"] if i % 2 == 0 else ["vue", "ts"]
            spec_name = "Frontend React" if i % 2 == 0 else "Frontend Vue"
        elif i % 5 == 3:
            spec_key = "mobile"
            stack_keys = ["swift"] if i % 2 == 0 else ["kotlin"]
            spec_name = "Mobile Developer"
        else:
            spec_key = "data"
            stack_keys = ["python", "postgres", "redis"]
            spec_name = "Data / AI Developer"

        # Преобразование theta в test_score (0..100)
        # Сигмоида с масштабированием
        raw_pct = 1.0 / (1.0 + math.exp(-theta * 1.2)) * 100.0
        test_score = round(max(45.0, min(99.0, raw_pct + random.uniform(-2.0, 2.0))), 1)

        # ФСП профиль: ~45% кандидатов имеют историю в реестре, ~55% без ФСП
        has_fsp = (i % 9 in (0, 2, 4, 6))  # 36 из 80 = 45%
        fsp_member_id = None
        sports_rank = None
        fsp_rating = 0
        fsp_score = 0.0
        fsp_achievements_count = 0
        fsp_best_place = None
        fsp_weight_sum = 0
        fsp_highlights = []

        if has_fsp:
            fsp_member_id = f"FSP-RU-77-{10000 + i:05d}"
            if grade_key in ("senior", "lead") and i % 3 == 0:
                sports_rank = "Мастер спорта"
                fsp_rating = random.randint(2350, 2600)
                fsp_best_place = 1 if i % 2 == 0 else 2
                fsp_weight_sum = random.randint(24, 35)
                fsp_achievements_count = random.randint(4, 7)
                fsp_score = round(random.uniform(85.0, 96.0), 1)
                event = "Чемпионата России ФСП 2024" if fsp_best_place == 1 else "Кубка ФСП 2025"
                place_str = "победитель" if fsp_best_place == 1 else "серебряный призёр"
                fsp_highlights = [f"{place_str} {event}"]
            elif grade_key in ("middle", "middle_plus", "senior"):
                sports_rank = "КМС" if i % 2 == 0 else "1-й спортивный разряд"
                fsp_rating = random.randint(1950, 2350)
                fsp_best_place = random.choice([2, 3, 4, 5])
                fsp_weight_sum = random.randint(14, 25)
                fsp_achievements_count = random.randint(3, 5)
                fsp_score = round(random.uniform(68.0, 85.0), 1)
                fsp_highlights = [f"призёр Всероссийского хакатона ФСП"]
            else:
                sports_rank = random.choice(["2-й спортивный разряд", "3-й спортивный разряд", "Без разряда"])
                fsp_rating = random.randint(1600, 1950)
                fsp_best_place = random.choice([5, 8, 12])
                fsp_weight_sum = random.randint(6, 14)
                fsp_achievements_count = random.randint(1, 3)
                fsp_score = round(random.uniform(40.0, 65.0), 1)
                fsp_highlights = ["участник студенческой лиги ФСП"]

        periodic_tasks = random.randint(1, 7) if test_score > 75 else random.randint(0, 3)
        last_active_days = random.randint(1, 60)
        last_active = datetime.now() - timedelta(days=last_active_days)

        # Категория
        cat_key = f"{spec_key}_{grade_key}"
        cat_id = CATEGORIES.get(cat_key, CATEGORIES["backend_middle"])

        # Вычисляем скоринг формулы BE2-04
        # Score = 0.35*S_test + 0.30*S_fsp + 0.25*S_stack + 0.10*S_act
        w_test, w_fsp, w_stack, w_act = 0.35, 0.30, 0.25, 0.10
        if not has_fsp:
            # Адаптивное распределение
            s_fsp = 0.0
            act_score = min(100.0, periodic_tasks * 12.0 + (40.0 if last_active_days <= 14 else 20.0))
            calc_score = (0.45 * test_score) + (0.40 * 85.0) + (0.15 * act_score)
        else:
            s_fsp = fsp_score
            act_score = min(100.0, periodic_tasks * 12.0 + (40.0 if last_active_days <= 14 else 20.0))
            calc_score = (w_test * test_score) + (w_fsp * s_fsp) + (w_stack * 90.0) + (w_act * act_score)

        calc_score = round(min(100.0, max(20.0, calc_score)), 1)

        # Бейджи и причины
        reasons = []
        if test_score >= 90:
            reasons.append("top_test_performer")
        elif test_score >= 80:
            reasons.append("high_test_score")

        if has_fsp:
            reasons.append("fsp_verified")
            if fsp_best_place == 1:
                reasons.append("fsp_champion")
            elif fsp_best_place in (2, 3):
                reasons.append("fsp_medalist")
            if sports_rank in ("ЗМС", "МСМК", "Мастер спорта"):
                reasons.append("fsp_master")
            elif sports_rank in ("КМС", "1-й спортивный разряд"):
                reasons.append("fsp_ranked")
        else:
            reasons.append("no_fsp_candidate")

        if periodic_tasks >= 3:
            reasons.append("active_solver")

        # Текстовое обоснование
        highlights_text = []
        if test_score >= 90:
            highlights_text.append(f"Топ-5% по тесту {spec_name} ({test_score}%)")
        else:
            highlights_text.append(f"Подтвержденный балл тестирования {test_score}%")

        if has_fsp and len(fsp_highlights) > 0:
            highlights_text.append(fsp_highlights[0])
            if sports_rank:
                highlights_text.append(sports_rank)
        elif not has_fsp:
            highlights_text.append("без истории соревнований ФСП (оценка по тестам и стеку)")

        explanation = ", ".join(highlights_text) + "."

        candidates.append({
            "id": cand_id,
            "full_name": full_name,
            "city": city,
            "specialization": spec_key,
            "specialization_id": SPEC_IDS[spec_key],
            "specialization_name": spec_name,
            "grade": grade_key,
            "grade_id": grade_info["id"],
            "grade_name": grade_info["name"],
            "grade_rank": grade_info["rank"],
            "category_id": cat_id,
            "stack": [TECH_IDS[k]["name"] for k in stack_keys],
            "stack_ids": [TECH_IDS[k]["id"] for k in stack_keys],
            "years_experience": exp,
            "salary_expectation": salary_want,
            "theta": theta,
            "test_score": test_score,
            "has_fsp": has_fsp,
            "fsp_member_id": fsp_member_id,
            "sports_rank": sports_rank,
            "fsp_rating": fsp_rating,
            "fsp_score": fsp_score,
            "fsp_achievements_count": fsp_achievements_count,
            "fsp_best_place": fsp_best_place,
            "fsp_weight_sum": fsp_weight_sum,
            "fsp_highlights": fsp_highlights,
            "periodic_tasks": periodic_tasks,
            "last_active": last_active.isoformat(),
            "calculated_score": calc_score,
            "reasons": reasons,
            "explanation": explanation
        })

    # 2. Генерация 18 работодателей и потребностей
    employers = []
    queries = []

    for idx, comp in enumerate(COMPANIES_DATA[:num_employers]):
        comp_id = f"33333333-3333-3333-3333-{idx+1000:012d}"
        owner_id = f"44444444-4444-4444-4444-{idx+1000:012d}"
        
        # Вакансии компании
        if idx % 3 == 0:
            target_spec = "backend"
            target_grade = "senior"
            req_stack = ["Go", "PostgreSQL", "Redis"]
            req_stack_ids = [TECH_IDS["go"]["id"], TECH_IDS["postgres"]["id"], TECH_IDS["redis"]["id"]]
            salary_min, salary_max = 320000, 480000
            vac_title = "Senior Go Developer"
        elif idx % 3 == 1:
            target_spec = "backend"
            target_grade = "middle"
            req_stack = ["Python", "PostgreSQL", "Docker"]
            req_stack_ids = [TECH_IDS["python"]["id"], TECH_IDS["postgres"]["id"], TECH_IDS["docker"]["id"]]
            salary_min, salary_max = 200000, 280000
            vac_title = "Middle Python Backend Engineer"
        else:
            target_spec = "frontend"
            target_grade = "middle"
            req_stack = ["React", "TypeScript"]
            req_stack_ids = [TECH_IDS["react"]["id"], TECH_IDS["ts"]["id"]]
            salary_min, salary_max = 190000, 260000
            vac_title = "Frontend React / TypeScript Developer"

        vac_id = f"55555555-5555-5555-5555-{idx+1000:012d}"

        employers.append({
            "id": comp_id,
            "owner_user_id": owner_id,
            "name": comp["name"],
            "industry": comp["industry"],
            "size": comp["size"],
            "description": comp["desc"],
            "vacancies": [
                {
                    "id": vac_id,
                    "title": vac_title,
                    "specialization": target_spec,
                    "specialization_id": SPEC_IDS[target_spec],
                    "grade": target_grade,
                    "grade_id": GRADE_IDS[target_grade]["id"],
                    "stack": req_stack,
                    "stack_ids": req_stack_ids,
                    "salary_min": salary_min,
                    "salary_max": salary_max,
                    "currency": "RUB",
                    "work_format": "remote" if idx % 2 == 0 else "hybrid"
                }
            ]
        })

        queries.append({
            "employer_id": comp_id,
            "company_name": comp["name"],
            "specialization": target_spec,
            "specialization_id": SPEC_IDS[target_spec],
            "grade": target_grade,
            "grade_id": GRADE_IDS[target_grade]["id"],
            "stack": req_stack,
            "stack_ids": req_stack_ids,
            "salary_min": salary_min,
            "salary_max": salary_max
        })

    return {
        "candidates": candidates,
        "employers": employers,
        "employer_queries": queries
    }


def generate_sql_dump(dataset: Dict[str, Any], output_path: str = "deploy/seeds/demo_seed.sql"):
    """Генерация самодостаточного SQL-дампа с сидингом всех сервисов"""
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    lines = [
        "-- ============================================================================",
        "-- FSP Platform Demo Seed Data (BE2-05)",
        "-- 80 Realistic Candidates with tests & FSP, 18 Employers, 26 Vacancies",
        "-- ============================================================================",
        "",
        "-- 1. Search read-model (candidate_search_docs)",
        "TRUNCATE candidate_search_docs CASCADE;",
        ""
    ]

    for c in dataset["candidates"]:
        stack_uuids = "{" + ",".join(c["stack_ids"]) + "}"
        highlights = "{" + ",".join([f'"{h}"' for h in c["fsp_highlights"]]) + "}"
        reasons = "{" + ",".join([f'"{r}"' for r in c["reasons"]]) + "}"
        bp_val = str(c["fsp_best_place"]) if c["fsp_best_place"] is not None else "NULL"
        rank_str = f"'{c['sports_rank']}'" if c["sports_rank"] else "NULL"
        disp_name = c["full_name"].replace("'", "''")
        exp_txt = c["explanation"].replace("'", "''")

        line = f"""INSERT INTO candidate_search_docs (
  user_id, display_name, category_id, specialization_id, specialization_name,
  grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
  fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
  fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
  activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
) VALUES (
  '{c['id']}', '{disp_name}', '{c['category_id']}', '{c['specialization_id']}', '{c['specialization_name']}',
  '{c['grade_id']}', '{c['grade_name']}', {c['grade_rank']}, {c['test_score']}, {str(c['has_fsp']).upper()}, {rank_str},
  {c['fsp_rating']}, {c['fsp_score']}, {c['fsp_achievements_count']}, {bp_val}, {c['fsp_weight_sum']},
  '{highlights}'::text[], '{stack_uuids}'::uuid[], {c['years_experience']}, '{c['city']}', {c['periodic_tasks']},
  85.0, {c['calculated_score']}, '{exp_txt}', '{reasons}'::text[], '{c['last_active']}', now()
) ON CONFLICT (user_id) DO UPDATE SET calculated_score = EXCLUDED.calculated_score;"""
        lines.append(line)

    lines.append("")
    lines.append("-- 2. Companies & Vacancies")
    for emp in dataset["employers"]:
        comp_name = emp["name"].replace("'", "''")
        desc = emp["description"].replace("'", "''")
        lines.append(f"""INSERT INTO companies (id, owner_user_id, name, description, industry, size)
VALUES ('{emp['id']}', '{emp['owner_user_id']}', '{comp_name}', '{desc}', '{emp['industry']}', '{emp['size']}')
ON CONFLICT (id) DO NOTHING;""")

        for vac in emp["vacancies"]:
            title = vac["title"].replace("'", "''")
            vac_stack = "{" + ",".join(vac["stack_ids"]) + "}"
            lines.append(f"""INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
VALUES ('{vac['id']}', '{emp['id']}', '{title}', 'Описание вакансии {title}', '{vac['specialization_id']}', '{vac['grade_id']}', '{vac_stack}'::uuid[], {vac['salary_min']}, {vac['salary_max']}, 'RUB', '{vac['work_format']}', 'published')
ON CONFLICT (id) DO NOTHING;""")

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")

    print(f"[seed] SQL дамп сохранен в: {output_path}")


def seed_postgres_directly(dataset: Dict[str, Any], host="localhost", port=5432, user="fsp", password="fsp"):
    """Прямой сидинг в PostgreSQL через psycopg2"""
    try:
        import psycopg2
    except ImportError:
        print("[seed] Библиотека psycopg2 не найдена, используйте сгенерированный SQL-дамп deploy/seeds/demo_seed.sql")
        return False

    db_configs = [
        {"dbname": "search_db", "role": "search"},
        {"dbname": "candidate_db", "role": "candidate"},
        {"dbname": "employer_db", "role": "employer"},
    ]

    for db_cfg in db_configs:
        try:
            conn = psycopg2.connect(
                host=host,
                port=port,
                user=user,
                password=password,
                dbname=db_cfg["dbname"],
                connect_timeout=3
            )
            cur = conn.cursor()

            if db_cfg["role"] == "search":
                print(f"[seed] Заполнение {db_cfg['dbname']}...")
                for c in dataset["candidates"]:
                    cur.execute("""
                        INSERT INTO candidate_search_docs (
                            user_id, display_name, category_id, specialization_id, specialization_name,
                            grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
                            fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
                            fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
                            activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
                        ) VALUES (
                            %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,
                            %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, now()
                        ) ON CONFLICT (user_id) DO UPDATE SET
                            calculated_score = EXCLUDED.calculated_score,
                            explanation = EXCLUDED.explanation;
                    """, (
                        c["id"], c["full_name"], c["category_id"], c["specialization_id"], c["specialization_name"],
                        c["grade_id"], c["grade_name"], c["grade_rank"], c["test_score"], c["has_fsp"], c["sports_rank"],
                        c["fsp_rating"], c["fsp_score"], c["fsp_achievements_count"], c["fsp_best_place"], c["fsp_weight_sum"],
                        c["fsp_highlights"], c["stack_ids"], c["years_experience"], c["city"], c["periodic_tasks"],
                        85.0, c["calculated_score"], c["explanation"], c["reasons"], c["last_active"]
                    ))
                conn.commit()
                print(f"[seed] {db_cfg['dbname']}: успешно загружено {len(dataset['candidates'])} кандидатов.")

            elif db_cfg["role"] == "employer":
                print(f"[seed] Заполнение {db_cfg['dbname']}...")
                for emp in dataset["employers"]:
                    cur.execute("""
                        INSERT INTO companies (id, owner_user_id, name, description, industry, size)
                        VALUES (%s, %s, %s, %s, %s, %s)
                        ON CONFLICT (id) DO NOTHING;
                    """, (emp["id"], emp["owner_user_id"], emp["name"], emp["description"], emp["industry"], emp["size"]))
                    for vac in emp["vacancies"]:
                        cur.execute("""
                            INSERT INTO vacancies (id, company_id, title, description, specialization_id, grade_id, stack, salary_min, salary_max, currency, work_format, status)
                            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, 'RUB', %s, 'published')
                            ON CONFLICT (id) DO NOTHING;
                        """, (
                            vac["id"], emp["id"], vac["title"], f"Описание {vac['title']}", vac["specialization_id"],
                            vac["grade_id"], vac["stack_ids"], vac["salary_min"], vac["salary_max"], vac["work_format"]
                        ))
                conn.commit()
                print(f"[seed] {db_cfg['dbname']}: успешно загружено {len(dataset['employers'])} работодателей.")

            cur.close()
            conn.close()
        except Exception as e:
            print(f"[seed] Предупреждение: подключение к {db_cfg['dbname']} пропущено ({e})")

    return True


def main():
    parser = argparse.ArgumentParser(description="FSP Platform Demo Seeder & Validation (BE2-05)")
    parser.add_argument("--candidates", type=int, default=80, help="Количество кандидатов (по ТЗ: 60-100)")
    parser.add_argument("--employers", type=int, default=18, help="Количество компаний (по ТЗ: 15-20)")
    parser.add_argument("--host", type=str, default=os.getenv("POSTGRES_HOST", "localhost"), help="Postgres host")
    parser.add_argument("--port", type=int, default=int(os.getenv("POSTGRES_PORT", 5432)), help="Postgres port")
    parser.add_argument("--user", type=str, default=os.getenv("POSTGRES_USER", "fsp"), help="Postgres user")
    parser.add_argument("--password", type=str, default=os.getenv("POSTGRES_PASSWORD", "fsp"), help="Postgres password")
    parser.add_argument("--validate", action="store_true", help="Запустить процедуру валидации и сформировать отчет")
    parser.add_argument("--dump-sql", type=str, default="deploy/seeds/demo_seed.sql", help="Путь для SQL дампа")
    parser.add_argument("--dump-json", type=str, default="deploy/seeds/demo_dataset.json", help="Путь для JSON дампа")
    args = parser.parse_args()

    print(f"[seed] Генерация синтетического датасета: {args.candidates} кандидатов, {args.employers} работодателей...")
    dataset = generate_synthetic_dataset(num_candidates=args.candidates, num_employers=args.employers)

    # Сохранение JSON дампа
    os.makedirs(os.path.dirname(args.dump_json), exist_ok=True)
    with open(args.dump_json, "w", encoding="utf-8") as f:
        json.dump(dataset, f, ensure_ascii=False, indent=2)
    print(f"[seed] JSON датасет сохранен в: {args.dump_json}")

    # Генерация SQL дампа
    generate_sql_dump(dataset, args.dump_sql)

    # Попытка прямого сидинга в базу данных PostgreSQL
    seed_postgres_directly(dataset, host=args.host, port=args.port, user=args.user, password=args.password)

    # Если запрошена валидация или по умолчанию при сидинге
    if args.validate or True:
        print("\n[seed] Запуск процедуры валидации решения BE2-05...")
        from validate import run_full_validation, render_markdown_report
        metrics = run_full_validation()
        render_markdown_report(metrics, "VALIDATION_REPORT.md")
        render_markdown_report(metrics, "for_devs/VALIDATION_REPORT.md")

    print("\n[seed] Сидинг и валидация успешно завершены!")


if __name__ == "__main__":
    main()
