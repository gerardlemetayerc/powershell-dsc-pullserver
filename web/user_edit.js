$(function() {
    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('id');
    if (!userId) {
        window.location.href = '/web/users';
        return;
    }
    // Charger les infos utilisateur
    $.ajax({
        url: '/api/v1/users/' + userId,
        type: 'GET',
        dataType: 'json',
        success: function(u) {
            console.log(u);
            $('#userId').val(u.id);
            $('#firstName').val(u.first_name);
            $('#lastName').val(u.last_name);
            $('#email').val(u.email);
            $('#isActive').val(u.is_active ? '1' : '0');
            $('#lastLogonDate').val(u.last_logon_date || '');
        }
    });
    // Sauvegarde
    $('#editUserForm').on('submit', function(e) {
        e.preventDefault();
        const data = {
            first_name: $('#firstName').val(),
            last_name: $('#lastName').val(),
            email: $('#email').val(),
            is_active: $('#isActive').val() === '1',
        };
        $.ajax({
            url: '/api/v1/users/' + userId,
            type: 'PUT',
            contentType: 'application/json',
            data: JSON.stringify(data),
            success: function() {
                window.location.href = '/web/users';
            }
        });
    });
});
