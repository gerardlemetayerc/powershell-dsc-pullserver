// Variable globale pour stocker l'ID de l'utilisateur
var currentUserId = null;

$(function() {
    // Récupère l'ID utilisateur depuis le JWT
    // L'email utilisateur est récupéré via l'API /api/v1/my

    // Récupère l'ID utilisateur via /api/v1/users?email=...
    $.get('/api/v1/my', function(user) {
        if (!user) return;
        console.log(user);
        $('#profileName').text(user.first_name + ' ' + user.last_name);
        userId = user.id;
        currentUserId = user.id; // Stocke l'ID utilisateur dans la variable globale
        loadTokens(user.id);
        // Création token
        $('#addTokenBtn').click(function() { $('#createTokenModal').modal('show'); });
        $('#createTokenForm').on('submit', function(e) {
            e.preventDefault();
            $.ajax({
                url: '/api/v1/users/' + user.id + '/tokens',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ label: $('#tokenLabel').val() }),
                success: function(resp) {
                    $('#tokenPlainDiv').text('Token à copier : ' + resp.token).show();
                    $('#createTokenModal').modal('hide');
                    loadTokens(user.id);
                },
                error: function() {
                    alert('Erreur création token');
                }
            });
        });
    });
    function loadTokens(userId) {
        $.get('/api/v1/users/' + userId + '/tokens', function(tokens) {
            // Utilise DataTables pour afficher les tokens
            if ($.fn.DataTable.isDataTable('#apiTokensTable')) {
                $('#apiTokensTable').DataTable().clear().rows.add(tokens).draw();
            } else {
                $('#apiTokensTable').DataTable({
                    data: tokens,
                    destroy: true,
                    columns: [
                        { data: 'label', defaultContent: '' },
                        { data: 'is_active', render: function(data) { return data ? 'Oui' : 'Non'; } },
                        { data: 'created_at', defaultContent: '' },
                        { data: 'revoked_at', defaultContent: '' },
                        { data: null, orderable: false, render: function(data, type, row) {
                            return (row.is_active ? '<button class="btn btn-sm btn-warning revoke-token" data-id="' + row.id + '">Révoquer</button> ' : '') +
                                   '<button class="btn btn-sm btn-danger delete-token" data-id="' + row.id + '">Supprimer</button>';
                        }}
                    ]
                });
            }
        });
    }
    // Actions révoquer/supprimer
    $('#apiTokensTable').on('click', '.revoke-token', function() {
        const id = $(this).data('id');
        $.post('/api/v1/users/' + currentUserId + '/tokens/' + id + '/revoke', function() {
            loadTokens(currentUserId);
        });
    });
    $('#apiTokensTable').on('click', '.delete-token', function() {
        const id = $(this).data('id');
        $.ajax({
            url: '/api/v1/users/' + currentUserId + '/tokens/' + id,
            type: 'DELETE',
            success: function() { loadTokens(currentUserId); }
        });
    });
});
