CREATE OR REPLACE FUNCTION category_children_json(p_tournament_id BIGINT, p_parent_id BIGINT)
RETURNS JSONB
LANGUAGE SQL
STABLE
AS $$
    SELECT COALESCE(
        jsonb_agg(
            to_jsonb(category) || jsonb_build_object(
                'children', category_children_json(p_tournament_id, category.id)
            )
            ORDER BY category.id DESC
        ),
        '[]'::JSONB
    )
    FROM categories AS category
    WHERE category.tournament_id = p_tournament_id
      AND category.parent_id = p_parent_id;
$$;
