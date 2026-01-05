$(document).ready(function() {
    // Table modules
    var table = $('#modules-table').DataTable({
        ajax: {
            url: '/api/v1/modules',
            dataSrc: ''
        },
        columns: [
            { data: 'name' },
            { data: 'version' },
            { data: 'checksum' },
            { data: 'uploaded_at' },
            {
                data: 'id',
                render: function(id, type, row) {
                    return `<button class="btn btn-danger btn-sm btn-delete-module" data-id="${id}"><i class="fas fa-trash"></i></button>`;
                },
                orderable: false,
                searchable: false
            }
        ]
    });
    // Bouton + ouvre la modale
    $('#btn-upload-module').on('click', function() {
        $('#modal-upload-module').modal('show');
    });
    // Formulaire upload
    $('#form-upload-module').on('submit', function(e) {
        e.preventDefault();
        var formData = new FormData(this);
        $.ajax({
            url: '/api/v1/modules/upload',
            type: 'POST',
            data: formData,
            processData: false,
            contentType: false,
            success: function(resp) {
                $('#modal-upload-module').modal('hide');
                table.ajax.reload();
                alert('Module uploadé avec succès !');
            },
            error: function(xhr) {
                alert('Erreur upload: ' + xhr.responseText);
            }
        });
    });
    // Suppression module avec confirmation
    $('#modules-table').on('click', '.btn-delete-module', function() {
        var id = $(this).data('id');
        if (confirm('Voulez-vous vraiment supprimer ce module ?')) {
            $.ajax({
                url: '/api/v1/modules/delete?id=' + id,
                type: 'DELETE',
                success: function(resp) {
                    table.ajax.reload();
                    alert('Module supprimé !');
                },
                error: function(xhr) {
                    alert('Erreur suppression: ' + xhr.responseText);
                }
            });
        }
    });
});
