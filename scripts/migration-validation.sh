BASE_BRANCH=$1

echo "BASE_BRANCH is: $BASE_BRANCH"

OLDEST_NEW_MIGRATION_FILE=$(git diff --name-only origin/$BASE_BRANCH --diff-filter=d | grep -m1 sql/postgres/migrations/)

if [[ -z $OLDEST_NEW_MIGRATION_FILE ]]; then
    echo "no new migrations"
    exit 0
fi

NEWEST_EXISTING_MIGRATION_FILE=$(git ls-tree -r origin/$BASE_BRANCH --name-only | grep sql/postgres/migrations/ | tail -1)

if [[ -z $NEWEST_EXISTING_MIGRATION_FILE ]]; then
    echo "no existing migrations"
    exit 0
fi

echo "oldest new migration $OLDEST_NEW_MIGRATION_FILE"
echo "newest existing migration $NEWEST_EXISTING_MIGRATION_FILE"

EXISTING_TIMESTAMP="$(basename $NEWEST_EXISTING_MIGRATION_FILE | cut -d '_' -f 1)"

NEW_TIMESTAMP="$(basename $OLDEST_NEW_MIGRATION_FILE | cut -d '_' -f 1)"

if [[ $EXISTING_TIMESTAMP -ge $NEW_TIMESTAMP ]]; then
    echo "existing migration timestamp is greater than or equal to incoming migration timestamp. please update your migrations timestamp."
    exit 1
fi

echo "new migration(s) are safe to merge"
exit 0
