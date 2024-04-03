ALTER TABLE adresses ADD COLUMN text_search tsvector;
UPDATE adresses
SET text_search = to_tsvector('french',
                              unaccent(nom) ||
                              ' ' || coalesce(codepost_4::text, ' ') ||
                              ' ' || coalesce(unaccent(localite), ' ') ||
                              ' ' || coalesce(unaccent(nom_com_of), ' ') ||
                              ' ' || coalesce(unaccent(voie), ' ') ||
                              ' ' || coalesce(unaccent(no_entree), ' '))
WHERE text_search IS NULL;
create index adresses_text_search_index on adresses using gin (text_search);

select nom,localite,voie,codepost_4,no_entree,text_search
from adresses
where text_search @@ plainto_tsquery('french', 'francois 14')
order by localite,voie,no_entree;