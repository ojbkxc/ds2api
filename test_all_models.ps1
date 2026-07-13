$baseUrl = "http://localhost:5001/v1/chat/completions"
$headers = @{
    "Authorization" = "Bearer your-api-key"
    "Content-Type" = "application/json"
}

$models = @(
    "deepseek-v4-flash",
    "deepseek-v4-flash-nothinking",
    "deepseek-v4-pro",
    "deepseek-v4-pro-nothinking",
    "deepseek-v4-vision",
    "deepseek-v4-vision-nothinking"
)

$testCases = @(
    @{ name = "Short message (below threshold)"; messages = '[{"role":"user","content":"Hi"}]'; enableCIF = $true; minChars = 100 },
    @{ name = "Long message (above threshold)"; messages = '[{"role":"user","content":"This is a very long message that exceeds the minimum character threshold for file upload. It contains enough text to trigger the current input file feature and test the upload functionality."}]'; enableCIF = $true; minChars = 0 },
    @{ name = "Multi-turn conversation"; messages = '[{"role":"user","content":"Hello"},{"role":"assistant","content":"Hi there!"},{"role":"user","content":"How are you?"}]'; enableCIF = $true; minChars = 0 },
    @{ name = "CIF disabled"; messages = '[{"role":"user","content":"Test with CIF disabled"}]'; enableCIF = $false; minChars = 0 }
)

foreach ($tc in $testCases) {
    Write-Host "`n=== Test Case: $($tc.name) ===" -ForegroundColor Cyan
    
    # Update config
    $config = @{
        keys = @("your-api-key")
        current_input_file = @{
            enabled = $tc.enableCIF
            min_chars = $tc.minChars
            filename_template = "{prefix}_{rand}.txt"
            disabled_models = @()
            vision_accounts = @()
        }
    } | ConvertTo-Json -Depth 5
    Set-Content -Path "C:\GitHub\ds2api\config.json" -Value $config -Encoding utf8
    
    # Wait for config reload
    Start-Sleep -Seconds 1
    
    foreach ($model in $models) {
        Write-Host "`nTesting model: $model" -ForegroundColor Yellow
        
        $body = @{
            model = $model
            messages = $tc.messages | ConvertFrom-Json
            stream = $false
        } | ConvertTo-Json -Depth 10
        
        try {
            $response = Invoke-RestMethod -Uri $baseUrl -Method Post -Body $body -Headers $headers -TimeoutSec 30
            Write-Host "SUCCESS: Model $model returned response with ID: $($response.id)" -ForegroundColor Green
        } catch {
            $errorMsg = $_.Exception.Message
            if ($errorMsg -match "rate_limit_exceeded|no accounts") {
                Write-Host "INFO: $model - $errorMsg" -ForegroundColor Gray
            } elseif ($errorMsg -match "authentication_failed") {
                Write-Host "INFO: $model - Authentication failed (expected with test credentials)" -ForegroundColor Gray
            } else {
                Write-Host "ERROR: $model - $errorMsg" -ForegroundColor Red
            }
        }
    }
}

Write-Host "`n=== All tests completed ===" -ForegroundColor Cyan
