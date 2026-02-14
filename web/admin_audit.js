$(document).ready(function() {
    // Menu admin audit
    $('#admin-menu').append('<li><a href="/web/admin/audit" id="admin-audit-link">Audit</a></li>');

    if(window.location.pathname === '/web/admin/audit') {
        $("#admin-content").html('<h2>Audit utilisateur</h2><div id="audit-table-block"><table id="audit-table" class="display table table-bordered table-sm" style="width:100%"><thead><tr><th>Email</th><th>Action</th><th>Cible</th><th>Détails</th><th>IP</th><th>Date</th></tr></thead></table></div>');
        $('#audit-table').DataTable({
            ajax: {
                url: '/api/v1/audit',
                dataSrc: ''
            },
            columns: [
                { data: 'UserEmail', defaultContent: '' },
                { data: 'Action' },
                { data: 'Target', defaultContent: '' },
                { data: 'Details', defaultContent: '' },
                { data: 'IPAddress', defaultContent: '' },
                { data: 'CreatedAt' }
            ],
            order: [[5, 'desc']],
        });
    }
});
