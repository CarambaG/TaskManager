const API_BASE = 'http://localhost:8080/api';

// Элементы DOM
const registerForm = document.getElementById('registerForm');
const loginForm = document.getElementById('loginForm');
const registerResult = document.getElementById('registerResult');
const loginResult = document.getElementById('loginResult');
const healthStatus = document.getElementById('healthStatus');
const statusIndicator = document.getElementById('statusIndicator');
const statusText = document.getElementById('statusText');
const logsContainer = document.getElementById('logs');

// Проверка здоровья системы
async function checkHealth() {
    try {
        console.log('Sending request to:', `${API_BASE}/health`);

        const response = await fetch(`${API_BASE}/health`);
        console.log('Response status:', response.status);

        if (response.ok) {
            const data = await response.json();
            console.log('Health check data:', data);
            statusIndicator.className = 'status-indicator status-online';
            statusText.textContent = `Система работает (${data.status})`;
        } else {
            console.log('Health check failed with status:', response.status);
            throw new Error(`HTTP error! status: ${response.status}`);
        }
    } catch (error) {
        console.error('Health check error:', error);
        statusIndicator.className = 'status-indicator status-offline';
        statusText.textContent = 'Сервер недоступен';
    }
}

// Регистрация пользователя
registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(registerForm);
    const userData = {
        login: formData.get('login'),
        password: formData.get('password')
    };


    try {
        const response = await fetch(`${API_BASE}/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(userData)
        });

        const result = await response.json();

        if (response.ok) {
            registerResult.className = 'result success';
            registerResult.textContent = `Пользователь ${result.login} успешно зарегистрирован!`;
            registerForm.reset();
        } else {
            throw new Error(result.error || 'Ошибка регистрации');
        }
    } catch (error) {
        registerResult.className = 'result error';
        registerResult.textContent = `Ошибка: ${error.message}`;
    }
});

// Вход пользователя
loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(loginForm);
    const userData = {
        login: formData.get('login'),
        password: formData.get('password')
    };

    try {
        const response = await fetch(`${API_BASE}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(userData)
        });

        const result = await response.json();

        if (response.ok) {
            // Сохраняем токен в localStorage
            localStorage.setItem('authToken', result.token);
            //localStorage.setItem('userId', result.id);

            loginResult.className = 'result success';
            loginResult.textContent = `Успешный вход! Добро пожаловать, ${result.login}!`;
            loginForm.reset();

            // Немедленный редирект на dashboard
            window.location.href = '/dashboard';
        } else {
            throw new Error(result.error || 'Ошибка входа');
        }
    } catch (error) {
        loginResult.className = 'result error';
        loginResult.textContent = `Ошибка: ${error.message}`;
    }
});

// Функция для получения заголовков с токеном
function getAuthHeaders() {
    const token = localStorage.getItem('authToken');
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
    };
}

// Пример защищенного запроса
async function makeProtectedRequest() {
    try {
        const response = await fetch(`${API_BASE}/protected`, {
            method: 'GET',
            headers: getAuthHeaders()
        });

        if (response.ok) {
            const data = await response.json();
            console.log('Protected data:', data);
        } else if (response.status === 401) {
            // Токен истек или невалиден
            localStorage.removeItem('authToken');
            window.location.href = '/'; // Перенаправляем на страницу входа
        }
    } catch (error) {
        console.error('Protected request error:', error);
    }
}

// Проверка аутентификации при загрузке страницы
function checkAuth() {
    const token = localStorage.getItem('authToken');
    if (token) {
        // Пользователь аутентифицирован
        // Можно обновить интерфейс
    }
}

// Валидация форм
function setupFormValidation() {
    const forms = [registerForm, loginForm];

    forms.forEach(form => {
        const inputs = form.querySelectorAll('input[required]');

        inputs.forEach(input => {
            input.addEventListener('input', () => {
                if (input.validity.valid) {
                    input.style.borderColor = '#28a745';
                } else {
                    input.style.borderColor = '#dc3545';
                }
            });
        });
    });
}

// Проверка, если пользователь уже вошел
function checkIfLoggedIn() {
    const token = localStorage.getItem('authToken');
    if (token) {
        // Если уже вошел, редиректим на dashboard
        window.location.href = '/dashboard';
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    setupFormValidation();
    checkHealth();
    checkIfLoggedIn(); // Проверяем, не вошел ли уже пользователь
});