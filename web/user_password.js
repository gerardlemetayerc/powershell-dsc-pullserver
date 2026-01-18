$(function() {
    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('id');
    if (!userId) {
        window.location.href = '/web/users';
        return;
    }
    $('#userId').val(userId);
    $('#passwordForm').on('submit', function(e) {
        e.preventDefault();
        const newPassword = $('#newPassword').val();
        if (!newPassword) return;
        $.ajax({
            url: '/api/v1/users/' + userId + '/password',
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ new_password: newPassword }),
            success: function() {
                $('#passwordMsg').text('Mot de passe changé avec succès.').addClass('text-success').removeClass('text-danger');
                $('#newPassword').val('');
            },
            error: function() {
                $('#passwordMsg').text('Erreur lors du changement de mot de passe.').addClass('text-danger').removeClass('text-success');
            }
        });
    });
});
