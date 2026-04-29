const API_BASE = '/api';

function validateNameField(name) {
    if (!name) return 'Name is required';
    if (name.length < 2) return 'Name must be at least 2 characters';
    if (name.length > 100) return 'Name must be less than 100 characters';
    const nameRegex = /^[a-zA-Z\s\-']+$/;
    if (!nameRegex.test(name)) return 'Name can only contain letters, spaces, hyphens, and apostrophes';
    return '';
}

function validateEmailField(email) {
    if (!email) return 'Email is required';
    if (email.length > 255) return 'Email must be less than 255 characters';
    const emailRegex = /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(email)) return 'Please enter a valid email address';
    return '';
}

function validateSubjectField(subject) {
    if (!subject) return 'Subject is required';
    if (subject.length < 3) return 'Subject must be at least 3 characters';
    if (subject.length > 200) return 'Subject must be less than 200 characters';
    return '';
}

function validateMessageField(message) {
    if (!message) return 'Message is required';
    if (message.length < 5) return 'Message must be at least 5 characters';
    if (message.length > 1000) return 'Message must be less than 1000 characters';
    return '';
}

function showFieldError(fieldId, errorId, errorMessage) {
    const field = document.getElementById(fieldId);
    const errorDiv = document.getElementById(errorId);
    if (errorMessage) {
        field.classList.add('error');
        errorDiv.textContent = errorMessage;
        errorDiv.classList.add('show');
    } else {
        field.classList.remove('error');
        errorDiv.textContent = '';
        errorDiv.classList.remove('show');
    }
}

document.getElementById('name').addEventListener('input', (e) => {
    const error = validateNameField(e.target.value);
    showFieldError('name', 'nameError', error);
});

document.getElementById('email').addEventListener('input', (e) => {
    const error = validateEmailField(e.target.value);
    showFieldError('email', 'emailError', error);
});

document.getElementById('subject').addEventListener('input', (e) => {
    const error = validateSubjectField(e.target.value);
    showFieldError('subject', 'subjectError', error);
});

document.getElementById('message').addEventListener('input', (e) => {
    const error = validateMessageField(e.target.value);
    showFieldError('message', 'messageError', error);
});

document.getElementById('feedbackForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const name = document.getElementById('name').value.trim();
    const email = document.getElementById('email').value.trim();
    const subject = document.getElementById('subject').value.trim();
    const message = document.getElementById('message').value.trim();
    
    const nameError = validateNameField(name);
    const emailError = validateEmailField(email);
    const subjectError = validateSubjectField(subject);
    const messageError = validateMessageField(message);
    
    if (nameError || emailError || subjectError || messageError) {
        showFieldError('name', 'nameError', nameError);
        showFieldError('email', 'emailError', emailError);
        showFieldError('subject', 'subjectError', subjectError);
        showFieldError('message', 'messageError', messageError);
        showMessage('Please fix the errors above', 'error');
        return;
    }
    
    const formData = { name, email, subject, message };
    
    try {
        const response = await fetch(API_BASE + '/feedback', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showMessage(result.message, 'success');
            document.getElementById('feedbackForm').reset();
            ['name', 'email', 'subject', 'message'].forEach(field => {
                showFieldError(field, field + 'Error', '');
            });
        } else {
            showMessage(result.error || 'Submission failed', 'error');
        }
    } catch (error) {
        showMessage('Error submitting form. Make sure the server is running on port 4000', 'error');
    }
});

function showMessage(msg, type) {
    const messageArea = document.getElementById('messageArea');
    messageArea.textContent = msg;
    messageArea.className = 'message ' + type;
    setTimeout(() => {
        messageArea.className = 'message';
    }, 5000);
}