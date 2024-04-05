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
where nom_com_of <> (select name from communes c where st_contains(c.geom, adresses.geom) limit 1)


