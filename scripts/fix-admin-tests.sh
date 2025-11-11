#!/bin/bash

# Script to fix adminapi tests to use CreateTestContext

FILES=(
    "internal/adminapi/nodes_test.go"
    "internal/adminapi/operators_test.go"
    "internal/adminapi/profiles_test.go"
    "internal/adminapi/sessions_test.go"
    "internal/adminapi/users_test.go"
)

for file in "${FILES[@]}"; do
    echo "Processing $file..."
    
    # Step 1: Change setupTestApp to capture return value
    sed -i.bak 's/setupTestApp(t, db)/appCtx := setupTestApp(t, db)/g' "$file"
    
    # Step 2: Change e.NewContext to CreateTestContext
    # This is more complex as we need to handle multi-line patterns
    # For now, do a simple replace that works for the common pattern
    perl -i -p0e 's/c := e\.NewContext\(req, rec\)/c := CreateTestContext(e, db, req, rec, appCtx)/g' "$file"
    
    echo "Fixed $file"
done

echo "Done! Please review changes and run tests."
