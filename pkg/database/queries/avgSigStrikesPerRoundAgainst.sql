
/*
first need fight IDs that fighter has fought in then find the other fighter's stats 
*/

SELECT SUM(sig_strikes_landed), fi.id, fi.first_name, fi.last_name 
FROM (
    SELECT fight_id, f.id, f.first_name, f.last_name 
    FROM fight_stats AS fs 
    JOIN fighters AS f ON fs.fighter_id = f.id
) AS fi
JOIN fight_stats AS fs ON fi.fight_id = fs.fight_id AND fi.id != fs.fighter_id 
GROUP BY fi.id 
ORDER BY 1;

