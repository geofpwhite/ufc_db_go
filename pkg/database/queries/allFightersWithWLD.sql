SELECT
    f.id,
    f.first_name,
    f.last_name,
    f.height,
    f.reach,
    SUM(CASE WHEN fs.outcome LIKE "%WIN%" THEN 1 ELSE 0 END) AS win_count,
    SUM(CASE WHEN fs.outcome LIKE "%LO%" THEN 1 ELSE 0 END) AS loss_count,
    SUM(CASE WHEN fs.outcome LIKE "%D%" THEN 1 ELSE 0 END) AS draw_count
FROM
    fighters AS f
    JOIN fight_stats AS fs ON f.id = fs.fighter_id
GROUP BY
    f.id, f.first_name, f.last_name

LIMIT 100 OFFSET $1;