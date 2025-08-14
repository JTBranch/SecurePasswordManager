#!/bin/bash

# Test Report Generator
# Generates an HTML report from test outputs in tmp/output

OUTPUT_DIR="tmp/output"
REPORT_FILE="$OUTPUT_DIR/test-report.html"

echo "Generating HTML test report..."

# Start HTML
cat > "$REPORT_FILE" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Test Execution Report</title>
    <meta charset="UTF-8">
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            margin: 0; 
            padding: 20px; 
            background-color: #f5f5f5; 
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: white; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header { 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
            color: white; 
            padding: 30px; 
            text-align: center; 
        }
        .header h1 { margin: 0; font-size: 2.5em; }
        .header p { margin: 10px 0 0 0; opacity: 0.9; }
        .section { 
            padding: 25px; 
            border-bottom: 1px solid #eee; 
        }
        .section:last-child { border-bottom: none; }
        .section h2 { 
            margin: 0 0 20px 0; 
            color: #333; 
            font-size: 1.5em;
            border-left: 4px solid #667eea;
            padding-left: 15px;
        }
        .test-result { 
            margin: 15px 0; 
            padding: 15px; 
            border-radius: 6px; 
            border-left: 5px solid #ddd;
        }
        .passed { 
            background: #f0f9f0; 
            border-left-color: #4CAF50; 
        }
        .failed { 
            background: #fef5f5; 
            border-left-color: #f44336; 
        }
        .warning { 
            background: #fffbf0; 
            border-left-color: #ff9800; 
        }
        .artifact-links { 
            margin: 20px 0; 
        }
        .artifact-links a { 
            display: inline-block; 
            margin: 5px 10px 5px 0; 
            padding: 8px 15px; 
            background: #667eea; 
            color: white; 
            text-decoration: none; 
            border-radius: 4px; 
            font-size: 0.9em;
            transition: background 0.3s;
        }
        .artifact-links a:hover { 
            background: #5a6fd8; 
        }
        .stats { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); 
            gap: 20px; 
            margin: 20px 0; 
        }
        .stat-card { 
            background: #f8f9fa; 
            padding: 20px; 
            border-radius: 6px; 
            text-align: center; 
        }
        .stat-number { 
            font-size: 2em; 
            font-weight: bold; 
            color: #667eea; 
        }
        .stat-label { 
            color: #666; 
            margin-top: 5px; 
        }
        .coverage-bar { 
            background: #e0e0e0; 
            border-radius: 10px; 
            overflow: hidden; 
            height: 20px; 
            margin: 10px 0; 
        }
        .coverage-fill { 
            height: 100%; 
            background: linear-gradient(90deg, #4CAF50, #45a049); 
            transition: width 0.3s; 
        }
        pre { 
            background: #f8f8f8; 
            padding: 15px; 
            border-radius: 4px; 
            overflow-x: auto; 
            font-size: 0.9em; 
        }
        .timestamp { 
            color: #666; 
            font-size: 0.9em; 
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 Test Execution Report</h1>
EOF

# Add timestamp and git info
echo "            <p class=\"timestamp\">Generated: $(date)</p>" >> "$REPORT_FILE"
if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
    echo "            <p class=\"timestamp\">Commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'N/A') | Branch: $(git branch --show-current 2>/dev/null || echo 'N/A')</p>" >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" << 'EOF'
        </div>

        <div class="section">
            <h2>📊 Test Statistics</h2>
            <div class="stats">
EOF

# Calculate test statistics
UNIT_TESTS_PASSED=0
UNIT_TESTS_FAILED=0
E2E_TESTS_PASSED=0
E2E_TESTS_FAILED=0
INTEGRATION_TESTS_PASSED=0
INTEGRATION_TESTS_FAILED=0

# Count unit test results
if [ -f "$OUTPUT_DIR/unit-test-results.json" ]; then
    UNIT_TESTS_PASSED=$(grep '"Action":"pass"' "$OUTPUT_DIR/unit-test-results.json" | grep '"Test":' | wc -l 2>/dev/null || echo "0")
    UNIT_TESTS_FAILED=$(grep '"Action":"fail"' "$OUTPUT_DIR/unit-test-results.json" | grep '"Test":' | wc -l 2>/dev/null || echo "0")
fi

# Count E2E test results
if [ -f "$OUTPUT_DIR/e2e-test-results.json" ]; then
    E2E_TESTS_PASSED=$(grep '"Action":"pass"' "$OUTPUT_DIR/e2e-test-results.json" | grep '"Test":' | wc -l 2>/dev/null || echo "0")
    E2E_TESTS_FAILED=$(grep '"Action":"fail"' "$OUTPUT_DIR/e2e-test-results.json" | grep '"Test":' | wc -l 2>/dev/null || echo "0")
fi

# Count integration test results  
if [ -f "$OUTPUT_DIR/integration-test-results.json" ]; then
    INTEGRATION_TESTS_PASSED=$(grep '"Action":"pass"' "$OUTPUT_DIR/integration-test-results.json" | grep '"Test":' | wc -l 2>/dev/null || echo "0")
    INTEGRATION_TESTS_FAILED=$(grep '"Action":"fail"' "$OUTPUT_DIR/integration-test-results.json" | grep '"Test":' | wc -l 2>/dev/null || echo "0")
fi

# Clean up whitespace from wc output and ensure we have numbers
UNIT_TESTS_PASSED=$(echo "$UNIT_TESTS_PASSED" | tr -d ' \n')
UNIT_TESTS_FAILED=$(echo "$UNIT_TESTS_FAILED" | tr -d ' \n')
E2E_TESTS_PASSED=$(echo "$E2E_TESTS_PASSED" | tr -d ' \n')
E2E_TESTS_FAILED=$(echo "$E2E_TESTS_FAILED" | tr -d ' \n')
INTEGRATION_TESTS_PASSED=$(echo "$INTEGRATION_TESTS_PASSED" | tr -d ' \n')
INTEGRATION_TESTS_FAILED=$(echo "$INTEGRATION_TESTS_FAILED" | tr -d ' \n')

TOTAL_PASSED=$((UNIT_TESTS_PASSED + E2E_TESTS_PASSED + INTEGRATION_TESTS_PASSED))
TOTAL_FAILED=$((UNIT_TESTS_FAILED + E2E_TESTS_FAILED + INTEGRATION_TESTS_FAILED))
TOTAL_TESTS=$((TOTAL_PASSED + TOTAL_FAILED))

# Add stats to HTML
cat >> "$REPORT_FILE" << EOF
                <div class="stat-card">
                    <div class="stat-number">$TOTAL_TESTS</div>
                    <div class="stat-label">Total Tests</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number" style="color: #4CAF50;">$TOTAL_PASSED</div>
                    <div class="stat-label">Passed</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number" style="color: #f44336;">$TOTAL_FAILED</div>
                    <div class="stat-label">Failed</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number" style="color: #ff9800;">$(ls "$OUTPUT_DIR"/*.json "$OUTPUT_DIR"/*.txt "$OUTPUT_DIR"/*.html "$OUTPUT_DIR"/*.out 2>/dev/null | wc -l)</div>
                    <div class="stat-label">Artifacts</div>
                </div>
EOF

cat >> "$REPORT_FILE" << 'EOF'
            </div>
        </div>

        <div class="section">
            <h2>📈 Coverage Report</h2>
EOF

# Add coverage information
if [ -f "$OUTPUT_DIR/coverage-summary.txt" ]; then
    COVERAGE=$(grep "Total Coverage:" "$OUTPUT_DIR/coverage-summary.txt" | awk '{print $3}' | sed 's/%//' 2>/dev/null || echo "0")
    
    cat >> "$REPORT_FILE" << EOF
            <div class="coverage-bar">
                <div class="coverage-fill" style="width: ${COVERAGE}%"></div>
            </div>
            <p>Coverage: <strong>${COVERAGE}%</strong></p>
            <div class="artifact-links">
                <a href="coverage.html">📊 Detailed Coverage Report</a>
                <a href="coverage.out">📄 Coverage Profile</a>
            </div>
EOF
else
    echo "            <p>Coverage information not available</p>" >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" << 'EOF'
        </div>

        <div class="section">
            <h2>🧪 Test Results</h2>
EOF

# Unit Tests
UNIT_STATUS="failed"
if [ "$UNIT_TESTS_FAILED" -eq 0 ]; then
    UNIT_STATUS="passed"
fi

cat >> "$REPORT_FILE" << EOF
            <div class="test-result $UNIT_STATUS">
                <h3>Unit Tests</h3>
                <p><strong>Passed:</strong> $UNIT_TESTS_PASSED | <strong>Failed:</strong> $UNIT_TESTS_FAILED</p>
                <div class="artifact-links">
                    <a href="unit-test-results.json">📄 JSON Results</a>
                </div>
            </div>
EOF

# Integration Tests  
INTEGRATION_STATUS="failed"
if [ "$INTEGRATION_TESTS_FAILED" -eq 0 ]; then
    INTEGRATION_STATUS="passed"
fi

cat >> "$REPORT_FILE" << EOF
            <div class="test-result $INTEGRATION_STATUS">
                <h3>Integration Tests</h3>
                <p><strong>Passed:</strong> $INTEGRATION_TESTS_PASSED | <strong>Failed:</strong> $INTEGRATION_TESTS_FAILED</p>
                <div class="artifact-links">
                    <a href="integration-test-results.json">📄 JSON Results</a>
                    <a href="integration-test-output.txt">📄 Full Output</a>
                </div>
            </div>
EOF

# E2E Tests
E2E_STATUS="failed"
if [ "$E2E_TESTS_FAILED" -eq 0 ]; then
    E2E_STATUS="passed"
fi

cat >> "$REPORT_FILE" << EOF
            <div class="test-result $E2E_STATUS">
                <h3>End-to-End Tests</h3>
                <p><strong>Passed:</strong> $E2E_TESTS_PASSED | <strong>Failed:</strong> $E2E_TESTS_FAILED</p>
                <div class="artifact-links">
                    <a href="e2e-test-results.json">📄 JSON Results</a>
                    <a href="e2e-test-output.txt">📄 Full Output</a>
EOF

# Add links to application logs if they exist
if [ -d "$OUTPUT_DIR/test-reports" ] && [ "$(ls -A "$OUTPUT_DIR/test-reports" 2>/dev/null)" ]; then
    echo "                    <a href=\"test-reports/\">📁 Application Logs</a>" >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" << 'EOF'
                </div>
            </div>
        </div>

        <div class="section">
            <h2>🔍 Code Quality</h2>
EOF

# Add code quality results
LINT_ISSUES=0
VET_ISSUES=0

if [ -f "$OUTPUT_DIR/lint-report.txt" ]; then
    LINT_ISSUES=$(wc -l < "$OUTPUT_DIR/lint-report.txt" 2>/dev/null || echo "0")
fi

if [ -f "$OUTPUT_DIR/vet-report.txt" ]; then
    VET_ISSUES=$(wc -l < "$OUTPUT_DIR/vet-report.txt" 2>/dev/null || echo "0")
fi

QUALITY_STATUS="warning"
if [ "$LINT_ISSUES" -eq 0 ] && [ "$VET_ISSUES" -eq 0 ]; then
    QUALITY_STATUS="passed"
fi

cat >> "$REPORT_FILE" << EOF
            <div class="test-result $QUALITY_STATUS">
                <h3>Code Quality Checks</h3>
                <p><strong>Go Vet Issues:</strong> $VET_ISSUES | <strong>Lint Issues:</strong> $LINT_ISSUES</p>
                <div class="artifact-links">
                    <a href="vet-report.txt">📄 Vet Report</a>
                    <a href="lint-report.txt">📄 Lint Report</a>
                </div>
            </div>
EOF

cat >> "$REPORT_FILE" << 'EOF'
        </div>

        <div class="section">
            <h2>📁 All Artifacts</h2>
            <div class="artifact-links">
EOF

# List all available artifacts
for ext in json txt html out log; do
    for file in "$OUTPUT_DIR"/*."$ext"; do
        if [ -f "$file" ]; then
            filename=$(basename "$file")
            echo "                <a href=\"$filename\">📄 $filename</a>" >> "$REPORT_FILE"
        fi
    done
done

cat >> "$REPORT_FILE" << 'EOF'
            </div>
        </div>
    </div>
</body>
</html>
EOF

echo "HTML test report generated: $REPORT_FILE"
echo "📊 Report summary:"
echo "   - Total tests: $TOTAL_TESTS"
echo "   - Passed: $TOTAL_PASSED"
echo "   - Failed: $TOTAL_FAILED"
if [ -f "$OUTPUT_DIR/coverage-summary.txt" ]; then
    echo "   - Coverage: $(grep "Total Coverage:" "$OUTPUT_DIR/coverage-summary.txt" | awk '{print $3}' 2>/dev/null || echo 'N/A')"
fi
