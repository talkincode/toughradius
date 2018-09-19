function hasPerms(session,perms){
    if(session.level==='super'){
        return true;
    }
    flag = false;
    for(var i in perms){
        if(session.perms.includes(perms[i])){
            flag = true;
            break;
        }
    }
    return flag;
}

var format_date = function (date,fmt) {
    var o = {
        "y+": date.getFullYear(),
        "M+": date.getMonth() + 1,                 //月份
        "d+": date.getDate(),                    //日
        "h+": date.getHours(),                   //小时
        "m+": date.getMinutes(),                 //分
        "s+": date.getSeconds(),                 //秒
        "q+": Math.floor((date.getMonth() + 3) / 3), //季度
        "S+": date.getMilliseconds()             //毫秒
    };
    for (var k in o) {
        if (new RegExp("(" + k + ")").test(fmt)){
            if(k === "y+"){
                fmt = fmt.replace(RegExp.$1, ("" + o[k]).substr(4 - RegExp.$1.length));
            }
            else if(k==="S+"){
                var lens = RegExp.$1.length;
                lens = lens===1?3:lens;
                fmt = fmt.replace(RegExp.$1, ("00" + o[k]).substr(("" + o[k]).length - 1,lens));
            }
            else{
                fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
            }
        }
    }
    return fmt;
};

function bytesToSize(bytes) {
    var sizes = ['bytes', 'K', 'M', 'G', 'T'];
    if (bytes == 0) return '0 Byte';
    var i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
    return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i];
}

function doPost(action,formValues){
    var form=document.createElement("FORM");
    document.body.appendChild(form);
    form.method = "post";
    form.action = action;
    form.style.display = "none";
    for(var k in formValues){
        if(k!=='submit'){
            var _input = document.createElement("input");
            _input.name=k;
            _input.value=formValues[k];
            form.appendChild(_input);
        }
    }
    form.submit();
}
function guid() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random()*16|0, v = c == 'x' ? r : (r&0x3|0x8);
        return v.toString(16);
    });
}
function isValidIP(ip) {
    var reg = /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/
    return reg.test(ip);
}

function isPhone(phone) {
    var reg=/^0?(1[0-9][0-9]|15[012356789]|18[0236789]|14[57])[0-9]{8}$/
    return reg.test(phone);
}

//判断输入的EMAIL格式是否正确
function isEmail(email)
{
    var reg=/^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$/;
    return reg.test(email);
}
