$(document).ready(function() {
    function loadProperties() {
        $.getJSON('/api/v1/properties', function(data) {
            var tbody = '';
            data.forEach(function(prop) {
                tbody += '<tr>' +
                    '<td>' + prop.id + '</td>' +
                    '<td>' + prop.name + '</td>' +
                    '<td>' + (prop.description || '') + '</td>' +
                    '<td>' + prop.priority + '</td>' +
                    '<td>' +
                        '<button class="btn btn-sm btn-info edit-btn" data-id="' + prop.id + '"><i class="fas fa-edit"></i></button> ' +
                        '<button class="btn btn-sm btn-danger delete-btn" data-id="' + prop.id + '"><i class="fas fa-trash"></i></button>' +
                    '</td>' +
                '</tr>';
            });
            $('#propertiesTable tbody').html(tbody);
        });
    }

    $(function() {
        loadProperties();

        $('#btnAddProperty').click(function() {
            $('#propertyId').val('');
            $('#propertyName').val('');
            $('#propertyDescription').val('');
            $('#propertyPriority').val('0');
            $('#propertyModal').modal('show');
        });

        $(document).on('click', '.edit-btn', function() {
            var id = $(this).data('id');
            $.getJSON('/api/v1/properties/' + id, function(prop) {
                $('#propertyId').val(prop.id);
                $('#propertyName').val(prop.name);
                $('#propertyDescription').val(prop.description);
                $('#propertyPriority').val(prop.priority);
                $('#propertyModal').modal('show');
            });
        });

        $(document).on('click', '.delete-btn', function() {
            if(confirm('Supprimer cette propriété ?')) {
                var id = $(this).data('id');
                $.ajax({
                    url: '/api/v1/properties/' + id,
                    type: 'DELETE',
                    success: function() { loadProperties(); }
                });
            }
        });

        $('#savePropertyBtn').click(function(e) {
            e.preventDefault();
            var id = $('#propertyId').val();
            var prop = {
                name: $('#propertyName').val(),
                description: $('#propertyDescription').val(),
                priority: parseInt($('#propertyPriority').val())
            };
            if(id) {
                $.ajax({
                    url: '/api/v1/properties/' + id,
                    type: 'PUT',
                    contentType: 'application/json',
                    data: JSON.stringify(prop),
                    success: function() {
                        $('#propertyModal').modal('hide');
                        loadProperties();
                    }
                });
            } else {
                $.ajax({
                    url: '/api/v1/properties',
                    type: 'POST',
                    contentType: 'application/json',
                    data: JSON.stringify(prop),
                    success: function() {
                        $('#propertyModal').modal('hide');
                        loadProperties();
                    }
                });
            }
        });
    });
});