/*
        createTable(toid, jsondata, check, edit, del)：用于动态创建table
        @toid：创建table到id为toid的节点下
        @jsondata：用于创建table的json格式的数据（key含表头标题）
        @check：是否创建查看按钮
        @edit：是否创建编辑按钮
        @del：是否创建删除按钮
*/
function createTable(toid, jsondata, check, edit, del) {
    var table = document.createElement("table");
    var trh, trc, td;

    for (i in jsondata) {
        //________________创建表头________________________________________
        if (i == 0) {
            trh = document.createElement("tr");
            for (j in jsondata[0]) { //根据数据在tr内创建td
                td = document.createElement("td");
                td.appendChild(document.createTextNode(j));
                td.style.background = "#C1DAD7"; //设置表头颜色
                trh.appendChild(td);
            }
            if (check == true) { //创建查看按钮
                td = document.createElement("td");
                td.appendChild(document.createTextNode("查看"));
                td.style.background = "#C1DAD7"; //设置表头颜色
                trh.appendChild(td);
            }
            if (edit == true) { //创建编辑按钮
                td = document.createElement("td");
                td.appendChild(document.createTextNode("编辑"));
                td.style.background = "#C1DAD7"; //设置表头颜色
                trh.appendChild(td);
            }
            if (del == true) { //创建删除按钮
                td = document.createElement("td");
                td.appendChild(document.createTextNode("删除"));
                td.style.background = "#C1DAD7"; //设置表头颜色
                trh.appendChild(td);
            }
            table.appendChild(trh);
        }
        //________________创建数据行________________________________________
        trc = document.createElement("tr"); //创建tr
        for (j in jsondata[i]) { //根据数据在tr内创建td
            td = document.createElement("td");
            td.appendChild(document.createTextNode(jsondata[i][j]));
            trc.appendChild(td);
        }
        if (check == true) { //创建查看按钮
            td = document.createElement("td");
            var btnCheck = document.createElement("button");
            btnCheck.appendChild(document.createTextNode("查看"));
            td.appendChild(btnCheck);
            trc.appendChild(td);
        }
        if (edit == true) { //创建编辑按钮
            td = document.createElement("td");
            var btnEdit = document.createElement("button");
            btnEdit.appendChild(document.createTextNode("编辑"));
            td.appendChild(btnEdit);
            trc.appendChild(td);
        }
        if (del == true) { //创建删除按钮
            td = document.createElement("td");
            var btnDel = document.createElement("button");
            btnDel.appendChild(document.createTextNode("删除"));
            td.appendChild(btnDel);
            trc.appendChild(td);
        }
        table.appendChild(trc);
    }
    document.getElementById(toid).appendChild(table);
}