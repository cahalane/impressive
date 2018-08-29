# Impressive

![](http://i.imgur.com/dXS1iob.png)

A library for translating UCC's Publish2 calendars into iCal files for use in Google Calendar or other calendar apps. Also, provides a webserver that performs this conversion.

## How do I use this?

A server running this tool is located on https://bigbertha.netsoc.co/impressive.

To use it, try this:

1. Go to https://mytimetable.ucc.ie and set up your modules. Leave the tab open and come back here.
2. Copy this code:  
    `t=JSON.parse(localStorage.getItem("scientia-session-authorization")),u=t.token_type+" "+t.access_token,jQuery.ajax({url:"https://scientia-api-2-1-0.azurewebsites.net//api/Calendar",headers:{Authorization:u},success:function(e){var t=JSON.stringify(e);jQuery.ajax({url:"https://bigbertha.netsoc.co/impressive/",data:t,method:"POST",success:function(e){var t=new Blob([e],{type:"text/calendar"});if(window.navigator.msSaveOrOpenBlob)window.navigator.msSaveOrOpenBlob(t,"calendar.ical");else{var a=document.createElement("a"),n=URL.createObjectURL(t);a.href=n,a.download="calendar.ical",document.body.appendChild(a),a.click(),setTimeout(function(){document.body.removeChild(a),window.URL.revokeObjectURL(n)},0)}}})}});`
3. Paste it in the Developer Console on the MyTimetable page.
4. With the file you just downloaded, go to the Calendars page on Google Calendar, click Import Calendar and select the file.
