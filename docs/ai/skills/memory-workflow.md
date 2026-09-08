# Skill: Memory Workflow

## Назначение

Управляет тем, что и куда сохранять в знания, и не даёт завершать итерацию без обновления состояния.

## Когда использовать

- появился новый подтверждённый вывод по проекту;
- появился вывод о локальном окружении пользователя;
- завершена значимая итерация;
- нужно решить, обновлять `AI_LEARN` или `AI_USER_LEARN`.

## Что покрывает

- разделение проектной и пользовательской памяти;
- нормализацию знаний вместо сырых логов;
- синхронизацию активного `system_state.md` без накопления истории;
- критерии завершённости итерации.

## Workflow

1. Классифицировать знание:
   - проектное и воспроизводимое — в `AI_LEARN`;
   - машинно-специфичное или приватное — в `AI_USER_LEARN`;
   - смешанное — разделить без дублирования.
2. Привязать знание к реальному артефакту: файлу, классу, конфигу, таблице, route, endpoint, job или подтверждённому запуску.
3. Сохранять не поток вывода, а короткое правило или факт, который уменьшает будущие ошибки.
4. После каждой значимой итерации синхронизировать релевантный `system_state.md`: оставить только активное состояние, риск и следующий шаг; не добавлять завершённую задачу как ещё одну историческую строку.
5. Если в итерации были новые или предотвращённые командные ошибки, обновлять и соответствующие `commands/*`.
6. Если тема разрастается, добавлять короткое резюме и ссылку из индекса, а не дублировать один и тот же текст в нескольких местах.
7. `current_iteration.md` заменять итогом последней итерации. Старый итог сохранять в тематическом файле, `AI_LEARN/archive/history/changelog.md` или `archive/`, если он ещё не зафиксирован.
8. Если стартовый файл стал журналом, сначала сохранить полный исторический снимок в `archive/`, затем собрать компактный актуальный файл.

## Новая сессия

В новой сессии считать память рабочим навигатором:

1. читать корневой маршрут;
2. открывать только релевантные записи;
3. идти по подтверждённым паттернам, а не переоткрывать уже известные правила;
4. при расхождении с кодом или последним подтверждённым запуском обновлять память.

## Критерии завершённости

Итерация завершена только если одновременно выполнено всё:

- задача выполнена или внешний блокер описан честно;
- изменения минимальны и проверены;
- релевантный `system_state.md` отражает текущее активное состояние;
- `current_iteration.md` содержит только последний итог;
- новые важные выводы сохранены;
- полезные `Failed` нормализованы в знания или явно помечены как непереиспользуемые.

## Чего не делать

- не сохранять секреты и локальные доступы в `AI_LEARN`;
- не создавать пустые записи ради схемы;
- не дублировать одни и те же знания в корне, skills и памяти;
- не сохранять гипотезу без привязки к артефакту;
- не дописывать историю завершённых задач в стартовые файлы;
- не считать итерацию завершённой без синхронизации состояния.

## Связанные файлы

- `AI_LEARN/index.md`
- `AI_LEARN/system_state.md`
- `AI_LEARN/project_overview.md`
- `AI_LEARN/current_iteration.md`
- `AI_LEARN/research/research_backlog.md`
- `AI_LEARN/research/coverage_map.md`
- `../AI_USER_LEARN/<project_name>/*`

# Skill: Memory Workflow

## Purpose

Manages what should be stored in knowledge, where it should be stored, and prevents an iteration from being considered complete without updating the current state.

## When to Use

- when a new confirmed project finding appears;
- when a finding about the user's local environment appears;
- when a significant iteration is completed;
- when it is necessary to decide whether to update `AI_LEARN` or `AI_USER_LEARN`.

## What It Covers

- separation of project and user-specific memory;
- normalization of knowledge instead of storing raw logs;
- synchronization of the active `system_state.md` without accumulating history;
- iteration completion criteria.

## Workflow

1. Classify the knowledge:
   - project-specific and reproducible — store it in `AI_LEARN`;
   - machine-specific or private — store it in `AI_USER_LEARN`;
   - mixed — split it without duplication.

2. Link the knowledge to a real artifact: a file, class, config, table, route, endpoint, job, or confirmed command execution.

3. Store not the raw output stream, but a short rule or fact that reduces the chance of future errors.

4. After each significant iteration, synchronize the relevant `system_state.md`: keep only the active state, risk, and next step; do not append a completed task as another historical entry.

5. If the iteration introduced new command errors or prevented previously known ones, update the corresponding `commands/*` entries as well.

6. If a topic grows too large, add a short summary and a link from the index instead of duplicating the same text across multiple places.

7. Replace `current_iteration.md` with the result of the latest iteration. Preserve the previous result in a thematic file, `AI_LEARN/archive/history/changelog.md`, or `archive/` if it has not already been recorded elsewhere.

8. If a startup file has turned into a log, first preserve the full historical snapshot in `archive/`, then rebuild a compact up-to-date version.

## New Session

In a new session, treat memory as a working navigator:

1. read the root route;
2. open only the relevant entries;
3. follow confirmed patterns instead of rediscovering already known rules;
4. if memory conflicts with the code or the latest confirmed execution, update the memory.

## Completion Criteria

An iteration is complete only if all of the following conditions are met:

- the task is completed or an external blocker is described honestly;
- changes are minimal and verified;
- the relevant `system_state.md` reflects the current active state;
- `current_iteration.md` contains only the latest result;
- new important findings are stored;
- useful `Failed` results are normalized into knowledge or explicitly marked as non-reusable.

## Do Not

- do not store secrets or local access details in `AI_LEARN`;
- do not create empty entries just to satisfy the structure;
- do not duplicate the same knowledge across the root, skills, and memory;
- do not store a hypothesis without linking it to an artifact;
- do not append completed task history to startup files;
- do not consider an iteration complete without synchronizing the state.

## Related Files

- `AI_LEARN/index.md`
- `AI_LEARN/system_state.md`
- `AI_LEARN/project_overview.md`
- `AI_LEARN/current_iteration.md`
- `AI_LEARN/research/research_backlog.md`
- `AI_LEARN/research/coverage_map.md`
- `../AI_USER_LEARN/<project_name>/*`
