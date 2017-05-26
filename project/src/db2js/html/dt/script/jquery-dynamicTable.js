/*
 * jquery-dynamicTable - 1.0.0
 * riancy
 * 844120@qq.com
 */

jQuery.fn.MakeTable = function (objColumn, objData, objClassName, RowClick) {

            //样式
            $(this).attr("class", objClassName);

            var sHtml = "";
            sHtml += "<thead>";

            var sTrHtml = "<tr>";
            $.each(objColumn, function (i) {

                sTrHtml += "<th ";
                sTrHtml += "style=\"width:" + objColumn[i].Width.toString() + "px\"";
                sTrHtml += ">";
                sTrHtml += objColumn[i].ColumnName;
                sTrHtml += "</th>";

            });
            sTrHtml += "</tr>";
            sHtml += sTrHtml + "</thead>";

            sHtml += "<tbody>";

            $.each(objData, function (i) {
                sTrHtml = "<tr";

                if (RowClick != null && RowClick != undefined) {
                    //alert(RowClick);
                    sTrHtml += " onclick=\"CheckRow(this," + RowClick + ")\"";
                }

                sTrHtml += ">";
                var objTr = objData[i];
                for (x in objTr) {
                    sTrHtml += "<td style=\"";

                    var objLinqData = jLinq.from(objColumn).equals("DataId", x).take()[0];
                    sTrHtml += "text-align:" + objLinqData.DataAlign + ";";
                    sTrHtml += "\" ";
                    if (objLinqData.OnClick != null) {
                        sTrHtml += " onclick=\"" + objLinqData.OnClick + "\"";
                    }

                    sTrHtml += ">";
                    if (objLinqData.Format != null) {
                        sTrHtml += objLinqData.Format(objTr[x]);
                    }
                    else {
                        sTrHtml += objTr[x];
                    }

                    sTrHtml += "</td>";
                    //alert(x);
                }
                // sTrHtml += objData[i].
                sTrHtml += "</tr>";
                sHtml += sTrHtml;
            });

            sHtml += "</tbody>";
            //alert(sHtml);
            // $("#" + sId + " > tbody:last").append(sTrHtml);
            var sId = this[0].id;
            $("#" + sId).append(sHtml);
        };

        jQuery.fn.TableBindData = function (objData, objConfig) {
            var sId = this[0].id;

            $.each(objData, function (i) {
                sTrHtml = "<tr>";
                var objTr = objData[i];
                for (x in objTr) {
                    sTrHtml += "<td style=\"";

                    sTrHtml += "text-align:" + jLinq.from(objConfig).equals("DataId", x).take()[0].DataAlign + ";";


                    sTrHtml += "\" >"
                    sTrHtml += objTr[x];
                    sTrHtml += "</td>";
                    //alert(x);
                }
                // sTrHtml += objData[i].
                sTrHtml += "</tr>";
                $("#" + sId + " > tbody:last").append(sTrHtml);

            });
        }
        
        
        function CheckRow(obj, fn) {

            var objTrList = $(obj).parent().parent().find("tbody>tr");

            $.each(objTrList, function (i) {
                $(objTrList[i]).attr("class", "");
            });

            $(obj).attr("class", "tdChecked");
            //;fn(obj);
        }