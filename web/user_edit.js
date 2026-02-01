$(function() {
    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('id');
    if (!userId) {
        window.location.href = '/web/users';
        return;
    }
    // Charger les infos utilisateur et la liste des rôles dynamiquement
    $.when(
        $.get('/api/v1/user_roles'),
        $.ajax({
            url: '/api/v1/users/' + userId,
            type: 'GET',
            dataType: 'json'
        })
    ).done(function(rolesResp, userResp) {
        var roles = rolesResp[0];
        var u = userResp[0];
        var $role = $('#role');
        $role.empty();
        $.each(roles, function(i, r) {
            $role.append($('<option>', { value: r, text: r }));
        });
        $('#userId').val(u.id);
        $('#firstName').val(u.first_name);
        $('#lastName').val(u.last_name);
        $('#email').val(u.email);
        $('#isActive').val(u.is_active ? '1' : '0');
        $('#lastLogonDate').val(u.last_logon_date || '');
        $role.val(u.role);
        // Masquer le changement de mot de passe si source != local
        if (u.source && u.source.toLowerCase() !== 'local') {
            $('#changePasswordSection').hide();
        } else {
            $('#changePasswordSection').show();
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
            role: $('#role').val()
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
