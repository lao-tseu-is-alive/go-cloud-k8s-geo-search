ALTER TABLE adresses ADD COLUMN text_search tsvector;
UPDATE adresses
SET text_search = to_tsvector('french',
                              coalesce(unaccent(nom), '') ||
                              ' ' || coalesce(codepost_4::text, ' ') ||
                              ' ' || coalesce(unaccent(localite), ' ') ||
                              ' ' || coalesce(unaccent(nom_com_of), ' ') ||
                              ' ' || coalesce(unaccent(voie), ' ') ||
                              ' ' || coalesce(unaccent(no_entree), ' '))
WHERE text_search IS NULL;
drop index adresses_text_search_index;
create index adresses_text_search_index on adresses using gin (text_search);
SELECT count(*) FROM adresses WHERE text_search IS NULL;



select nom,localite,codepost_4,nom_com_of, voie,no_entree,
       (select name from communes c where st_contains(c.geom, adresses.geom) limit 1) as commune_geom,
       text_search
from adresses
where nom_com_of <> (select name from communes c where st_contains(c.geom, adresses.geom) limit 1)
--where text_search @@ plainto_tsquery('french', 'poste monnaz ')
order by localite,voie,no_entree;


select nom,localite,codepost_4,nom_com_of, voie,no_entree
from adresses
where nom_com_of <> (select name from communes c where st_contains(c.geom, adresses.geom) limit 1);

SELECT
            ROW_NUMBER() OVER (ORDER BY keywords)            AS id,
            subject,
            keywords,
            display,
            x,y,created_at
INTO search_item
FROM (select
          'adresse'                                        as subject,
          coalesce(codepost_4::text, '') ||
          ' ' || coalesce(lower(unaccent(nom_com_of)), ' ') ||
          ', ' || coalesce(unaccent(nom) || ', ', '')
              --    || coalesce(lower(split_part(unaccent(voie), ',', 1)), ' ') ||
              || coalesce(lower(unaccent(voie)), ' ') ||
          ' ' || coalesce(lower(unaccent(no_entree)), ' ') as keywords,
          coalesce(nom || ', ', '') || coalesce(voie_txt, '') ||
          ' ' || coalesce(lower(no_entree), '') ||
          ', ' || coalesce(codepost_4::text, '') ||
          ' ' || coalesce(nom_com_of, ' ')                 as display,
          --min(st_x(geom))::text || '_' || min(st_y(geom))::text as position
          round(min(st_x(geom)))::integer                  as x,
          round(min(st_y(geom)))::integer                  as y,
          now()::timestamp                                 as created_at
      FROM adresses
      GROUP BY keywords, display
      ORDER BY keywords asc
) AS items;


SELECT count(*) ,keywords FROM search_item GROUP BY keywords
HAVING count(*) > 1
ORDER BY count(*) desc, keywords asc;
