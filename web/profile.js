// Variable globale pour stocker l'ID de l'utilisateur
var currentUserId = null;

async function copyTextToClipboard(text) {
    if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        return;
    }

    const tempInput = document.createElement('textarea');
    tempInput.value = text;
    tempInput.setAttribute('readonly', '');
    tempInput.style.position = 'absolute';
    tempInput.style.left = '-9999px';
    document.body.appendChild(tempInput);
    tempInput.select();
    document.execCommand('copy');
    document.body.removeChild(tempInput);
}

$(function() {
    // Récupère l'ID utilisateur depuis le JWT
    // The user's email is retrieved via the /api/v1/my API endpoint, which returns the current user's information based on the JWT token sent in the request headers. This allows us to get the user's ID without needing to decode the JWT on the client side.

    // Retrieve the user ID via /api/v1/users?email=...
    $.get('/api/v1/my', function(user) {
        if (!user) return;
        console.log(user);
        $('#profileName').text(user.first_name + ' ' + user.last_name);
        userId = user.id;
        currentUserId = user.id; // Store the user ID in the global variable
        loadTokens(user.id);
        // Create token
        $('#addTokenBtn').click(function() { $('#createTokenModal').modal('show'); });
        $('#createTokenForm').on('submit', function(e) {
            e.preventDefault();
            $.ajax({
                url: '/api/v1/users/' + user.id + '/tokens',
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ label: $('#tokenLabel').val() }),
                success: function(resp) {
                    $('#tokenPlainCode').text(resp.token);
                    $('#copyTokenFeedback').text('Copy this token now. It will not be shown again.');
                    $('#tokenPlainDiv').show();
                    $('#createTokenModal').modal('hide');
                    $('#tokenLabel').val('');
                    loadTokens(user.id);
                },
                error: function() {
                    alert('Error creating token');
                }
            });
        });
    });
    function loadTokens(userId) {
        $.get('/api/v1/users/' + userId + '/tokens', function(tokens) {
            // Use DataTables to display tokens
            if ($.fn.DataTable.isDataTable('#apiTokensTable')) {
                $('#apiTokensTable').DataTable().clear().rows.add(tokens).draw();
            } else {
                $('#apiTokensTable').DataTable({
                    data: tokens,
                    destroy: true,
                    columns: [
                        { data: 'label', defaultContent: '' },
                        { data: 'is_active', render: function(data) { return data ? 'Yes' : 'No'; } },
                        { data: 'created_at', defaultContent: '' },
                        { data: 'revoked_at', defaultContent: '' },
                        { data: null, orderable: false, render: function(data, type, row) {
                            return (row.is_active ? '<button class="btn btn-sm btn-warning revoke-token" data-id="' + row.id + '">Revoke</button> ' : '') +
                                   '<button class="btn btn-sm btn-danger delete-token" data-id="' + row.id + '">Delete</button>';
                        }}
                    ]
                });
            }
        });
    }
    // Actions revoke/delete
    $('#apiTokensTable').on('click', '.revoke-token', function() {
        const id = $(this).data('id');
        if (!window.confirm('Revoke this API token? It will stop working immediately.')) {
            return;
        }
        $.post('/api/v1/users/' + currentUserId + '/tokens/' + id + '/revoke', function() {
            loadTokens(currentUserId);
        }).fail(function(xhr) {
            alert(xhr.responseText || 'Error revoking token');
        });
    });
    $('#apiTokensTable').on('click', '.delete-token', function() {
        const id = $(this).data('id');
        if (!window.confirm('Delete this API token permanently?')) {
            return;
        }
        $.ajax({
            url: '/api/v1/users/' + currentUserId + '/tokens/' + id,
            type: 'DELETE',
            success: function() { loadTokens(currentUserId); },
            error: function(xhr) {
                alert(xhr.responseText || 'Error deleting token');
            }
        });
    });

    $('#copyTokenBtn').on('click', async function() {
        const token = $('#tokenPlainCode').text();
        if (!token) {
            return;
        }

        try {
            await copyTextToClipboard(token);
            $('#copyTokenFeedback').text('Token copied to clipboard.');
        } catch (error) {
            $('#copyTokenFeedback').text('Copy failed. Select and copy the token manually.');
        }
    });
});
