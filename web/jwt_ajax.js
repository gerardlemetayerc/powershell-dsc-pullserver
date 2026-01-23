// Ajoute automatiquement le header Authorization: Bearer <token> à toutes les requêtes AJAX jQuery
$.ajaxSetup({
    beforeSend: function(xhr) {
        const token = localStorage.getItem('jwt_token');
        if (token) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + token);
        }
    }
});
