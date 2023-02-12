# Detect API enable status, API specific account status
:if ([/ip service print count-only where disabled=no name=api]=0) do {
   /ip service enable api;
   /ip service set api address=10.10.10.0/24
   log info message="enable api";
}

:if ([/user print count-only where name=apimaster]=0) do {
   /user add name=apimaster password="Api.2023!" group=write;
   log info message="add api account";
}

:if ([/user print count-only where name=apimaster disabled=no]=0) do {
   user set apimaster disabled=no;
   log info message="enable api account";
}