# PowerShell script to deploy all services to Cloud Run using their .env.gcp files
$services = @(
    @{ Name = "orders";   EnvFile = ".\orders\.env.gcp";   Image = "us-central1-docker.pkg.dev/lucky-dahlia-466206-k0/docker-repo/orders:latest" },
    @{ Name = "products"; EnvFile = ".\products\.env.gcp"; Image = "us-central1-docker.pkg.dev/lucky-dahlia-466206-k0/docker-repo/products:latest" },
    @{ Name = "payment";  EnvFile = ".\payment\.env.gcp";  Image = "us-central1-docker.pkg.dev/lucky-dahlia-466206-k0/docker-repo/payment:latest" },
    @{ Name = "authentication"; EnvFile = ".\authentication\.env.gcp"; Image = "us-central1-docker.pkg.dev/lucky-dahlia-466206-k0/docker-repo/authentication:latest" }
)
$region = "us-central1"

foreach ($svc in $services) {
    Write-Host "Deploying $($svc.Name)..."
    $envVars = Get-Content $svc.EnvFile | Where-Object { $_ -and $_ -notmatch '^#' } | ForEach-Object {
        $name, $value = $_ -split '=', 2
        "$name=$value"
    }
    $envVarsString = $envVars -join ','
    # Add your Cloud SQL instance connection name here
    $cloudSqlInstance = "lucky-dahlia-466206-k0:us-central1:gcp-tech-ready-level-2"
    gcloud run deploy $svc.Name `
        --image $svc.Image `
        --region $region `
        --set-env-vars $envVarsString `
        --add-cloudsql-instances $cloudSqlInstance `
        --platform managed
}
