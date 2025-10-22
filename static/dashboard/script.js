const API_BASE = 'http://localhost:8080/api';

// Элементы DOM
let currentUser = null;

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', async () => {
    // Сначала проверяем аутентификацию
    const isAuthenticated = await checkAuth();

    if (isAuthenticated) {
        // Если аутентифицирован, загружаем остальные данные
        await loadUserData();
        await loadTasks();
        setupEventListeners();
       //addLog('Панель управления загружена', 'success');
    }
});

// Проверка аутентификации
async function checkAuth() {
    const token = localStorage.getItem('authToken');
    if (!token) {
        // Если токена нет, редиректим на главную страницу
        window.location.href = '/';
        return false;
    }

    try {
        const response = await fetch(`${API_BASE}/me`, {
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Not authenticated');
        }

        currentUser = await response.json();
        updateUI();
        return true;
    } catch (error) {
        console.error('Auth check failed:', error);
        // Очищаем невалидный токен и редиректим
        localStorage.removeItem('authToken');
        localStorage.removeItem('userId');
        localStorage.removeItem('userLogin');
        window.location.href = '/';
        return false;
    }
}

// Загрузка данных пользователя
async function loadUserData() {
    try {
        const response = await fetch(`${API_BASE}/me`, {
            headers: getAuthHeaders()
        });

        if (response.ok) {
            const userData = await response.json();
            displayUserData(userData);
        }
    } catch (error) {
        console.error('Failed to load user data:', error);
        showNotification('Ошибка загрузки данных', 'error');
    }
}

// Отображение данных пользователя
function displayUserData(user) {
    // Обновление заголовка
    document.getElementById('userGreeting').textContent = `Добро пожаловать, ${user.login}!`;

    // Обновление карточки пользователя
    document.getElementById('userInitials').textContent = user.login.charAt(0).toUpperCase();
    document.getElementById('userLogin').textContent = user.login;
    document.getElementById('userSince').textContent = `Зарегистрирован: ${new Date(user.create_at).toLocaleDateString()}`;

    // Обновление профиля
    document.getElementById('profileUserId').textContent = user.id;
    document.getElementById('profileLogin').textContent = user.login;
    document.getElementById('profileCreateDate').textContent = new Date(user.create_at).toLocaleDateString();
}

// Загрузка задач
async function loadTasks() {
    try {
        const response = await fetch(`${API_BASE}/tasks`, {
            headers: getAuthHeaders()
        });

        if (response.ok) {
            const tasks = await response.json();
            displayTasks(tasks);
            updateStats(tasks);
        }
    } catch (error) {
        console.error('Failed to load tasks:', error);
        showNotification('Ошибка загрузки задач', 'error');
    }
}

// Отображение задач
function displayTasks(tasks) {
    const tasksList = document.getElementById('tasksList');

    if (!tasks || tasks.length === 0) {
        tasksList.innerHTML = `
            <div class="empty-state">
                <p>Задачи отсутствуют</p>
                <button class="btn-primary" onclick="openTaskModal()">
                    Создать первую задачу
                </button>
            </div>
        `;
        return;
    }

    tasksList.innerHTML = tasks.map(task => `
        <div class="task-item" data-task-id="${task.id}" data-status="${task.status}">
            <div class="task-header">
                <div class="task-title">${escapeHtml(task.title)}</div>
                <div class="task-priority priority-${task.priority}">
                    ${getPriorityText(task.priority)}
                </div>
            </div>
            ${task.description ? `<div class="task-description">${escapeHtml(task.description)}</div>` : ''}
            <div class="task-footer">
                <div class="task-meta">
                    ${task.due_date ? `Срок: ${new Date(task.due_date).toLocaleDateString()}` : 'Без срока'}
                </div>
                <div class="task-actions">
                    <button class="task-action-btn" onclick="toggleTaskStatus('${task.id}')" title="${task.status === 'completed' ? 'Вернуть в работу' : 'Завершить'}">
                        ${task.status === 'completed' ? '↶' : '✓'}
                    </button>
                    <button class="task-action-btn" onclick="editTask('${task.id}')" title="Редактировать">
                        ✏️
                    </button>
                    <button class="task-action-btn" onclick="deleteTask('${task.id}')" title="Удалить">
                        🗑️
                    </button>
                </div>
            </div>
        </div>
    `).join('');
}

// Обновление статистики
function updateStats(tasks) {
    const totalTasks = tasks.length;
    const completedTasks = tasks.filter(task => task.status === 'completed').length;
    const activeTasks = totalTasks - completedTasks;
    const todayTasks = tasks.filter(task => {
        const today = new Date().toDateString();
        const taskDate = new Date(task.due_date).toDateString();
        return taskDate === today;
    }).length;

    // Обновление профиля
    document.getElementById('totalTasks').textContent = totalTasks;
    document.getElementById('activeTasks').textContent = activeTasks;
    document.getElementById('completedTasks').textContent = completedTasks;

    // Обновление статистики
    document.getElementById('statTotalTasks').textContent = totalTasks;
    document.getElementById('statPendingTasks').textContent = activeTasks;
    document.getElementById('statCompletedTasks').textContent = completedTasks;
    document.getElementById('statTodayTasks').textContent = todayTasks;
}

// Создание новой задачи
async function createTask(taskData) {
    try {
        const response = await fetch(`${API_BASE}/tasks`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(taskData)
        });

        if (response.ok) {
            showNotification('Задача успешно создана', 'success');
            closeTaskModal();
            await loadTasks();
        } else {
            throw new Error('Failed to create task');
        }
    } catch (error) {
        console.error('Failed to create task:', error);
        showNotification('Ошибка создания задачи', 'error');
    }
}

// Переключение статуса задачи
async function toggleTaskStatus(taskId) {
    try {
        const response = await fetch(`${API_BASE}/tasks/${taskId}/toggle`, {
            method: 'PUT',
            headers: getAuthHeaders()
        });

        if (response.ok) {
            await loadTasks();
        } else {
            throw new Error('Failed to toggle task status');
        }
    } catch (error) {
        console.error('Failed to toggle task status:', error);
        showNotification('Ошибка обновления задачи', 'error');
    }
}

// Удаление задачи
async function deleteTask(taskId) {
    if (!confirm('Вы уверены, что хотите удалить эту задачу?')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/tasks/${taskId}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        if (response.ok) {
            showNotification('Задача удалена', 'success');
            await loadTasks();
        } else {
            throw new Error('Failed to delete task');
        }
    } catch (error) {
        console.error('Failed to delete task:', error);
        showNotification('Ошибка удаления задачи', 'error');
    }
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Форма создания задачи
    document.getElementById('taskForm').addEventListener('submit', async (e) => {
        e.preventDefault();

        const taskData = {
            title: document.getElementById('taskTitle').value,
            description: document.getElementById('taskDescription').value,
            priority: document.getElementById('taskPriority').value,
            due_date: document.getElementById('taskDueDate').value || null
        };

        await createTask(taskData);
    });

    // Фильтрация задач
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');

            const filter = btn.dataset.filter;
            filterTasks(filter);
        });
    });

    // Поиск задач
    document.getElementById('taskSearch').addEventListener('input', (e) => {
        searchTasks(e.target.value);
    });
}

// Фильтрация задач
function filterTasks(filter) {
    const tasks = document.querySelectorAll('.task-item');

    tasks.forEach(task => {
        switch (filter) {
            case 'active':
                task.style.display = task.dataset.status === 'active' ? 'block' : 'none';
                break;
            case 'completed':
                task.style.display = task.dataset.status === 'completed' ? 'block' : 'none';
                break;
            default:
                task.style.display = 'block';
        }
    });
}

// Поиск задач
function searchTasks(query) {
    const tasks = document.querySelectorAll('.task-item');
    const searchTerm = query.toLowerCase();

    tasks.forEach(task => {
        const title = task.querySelector('.task-title').textContent.toLowerCase();
        const description = task.querySelector('.task-description')?.textContent.toLowerCase() || '';

        if (title.includes(searchTerm) || description.includes(searchTerm)) {
            task.style.display = 'block';
        } else {
            task.style.display = 'none';
        }
    });
}

// Навигация по секциям
function showSection(sectionName) {
    // Скрыть все секции
    document.querySelectorAll('.content-section').forEach(section => {
        section.classList.remove('active');
    });

    // Показать выбранную секцию
    document.getElementById(`${sectionName}-section`).classList.add('active');

    // Обновить активную кнопку навигации
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
}

// Управление модальным окном
function openTaskModal() {
    document.getElementById('taskModal').style.display = 'block';
}

function closeTaskModal() {
    document.getElementById('taskModal').style.display = 'none';
    document.getElementById('taskForm').reset();
}

// Выход из системы
function logout() {
    localStorage.removeItem('authToken');
    localStorage.removeItem('userId');
    localStorage.removeItem('userLogin');
    window.location.href = '/';
}

// Вспомогательные функции
function getAuthHeaders() {
    const token = localStorage.getItem('authToken');
    return {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
    };
}

function getPriorityText(priority) {
    const priorities = {
        'low': 'Низкий',
        'medium': 'Средний',
        'high': 'Высокий'
    };
    return priorities[priority] || priority;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showNotification(message, type = 'info') {
    // Создаем уведомление
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;

    // Стили для уведомления
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px 20px;
        border-radius: 6px;
        color: white;
        z-index: 10000;
        font-weight: 500;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    `;

    if (type === 'success') {
        notification.style.background = '#28a745';
    } else if (type === 'error') {
        notification.style.background = '#dc3545';
    } else {
        notification.style.background = '#17a2b8';
    }

    document.body.appendChild(notification);

    // Автоматическое удаление через 5 секунд
    setTimeout(() => {
        notification.remove();
    }, 5000);
}

// Обновление интерфейса
function updateUI() {
    const userLogin = localStorage.getItem('userLogin');
    if (userLogin) {
        document.getElementById('userGreeting').textContent = `Добро пожаловать, ${userLogin}!`;
    }
}

// Закрытие модального окна при клике вне его
window.addEventListener('click', (e) => {
    const modal = document.getElementById('taskModal');
    if (e.target === modal) {
        closeTaskModal();
    }
});