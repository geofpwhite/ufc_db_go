SELECT 
    SUM(rs.sig_strikes_landed) / COUNT(rs.id) AS avground,
    f.first_name,
    f.last_name,
    COUNT(fs.fighter_id)
FROM 
    round_stats AS rs
JOIN 
    fight_stats AS fs 
    ON fs.id = rs.fight_stats_id
JOIN 
    fighters AS f 
    ON f.id = fs.fighter_id
    
WHERE
    f.id = $1
GROUP BY 
    f.id
ORDER BY 
    SUM(f.id);