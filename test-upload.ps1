$headers = @{
    'Tus-Resumable' = '1.0.0'
    'Upload-Length' = '1000'
}

try {
    $response = Invoke-WebRequest -Uri 'http://localhost:5000/api/v1/upload/files' -Method POST -Headers $headers -UseBasicParsing
    Write-Host "Status Code: $($response.StatusCode)"
    
    if ($response.Headers.ContainsKey('Location')) {
        Write-Host "Location Header: $($response.Headers['Location'])"
    } else {
        Write-Host "No Location header found"
    }
} catch {
    Write-Host "Error: $_"
}
