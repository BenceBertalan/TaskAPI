
$task = @{"ID"=0;"Name"="TestTask";"Command"="dir C:\Windows";Output=""}

$task_json = $task | ConvertTo-Json #| % { [System.Text.RegularExpressions.Regex]::Unescape($_) } 

$wscall = Invoke-WebRequest -Uri http://localhost:3000/api/tasks -Method Post -Body $task_json

$wscall = Invoke-WebRequest -Uri http://192.168.11.210:3000/api/tasks -Method Patch -Body $task_json

(Invoke-WebRequest -Uri http://localhost:3000/api/tasks).Content | ConvertFrom-Json