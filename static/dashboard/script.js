const API_BASE = 'http://localhost:8080/api';

// Элементы DOM
let currentUser = null;
let currentEditingTaskId = null;

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', async () => {
    // Сначала проверяем аутентификацию
    const isAuthenticated = await checkAuth();

    if (isAuthenticated) {
        // Если аутентифицирован, загружаем остальные данные
        await loadUserData();
        await loadTasks();
        setupEventListeners();
    }
});

// Проверка аутентификации
async function checkAuth() {
    const token = localStorage.getItem('authToken');
    if (!token) {
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
    document.getElementById('userGreeting').textContent = `Добро пожаловать, ${user.login}!`;
    document.getElementById('userInitials').textContent = user.login.charAt(0).toUpperCase();
    document.getElementById('userLogin').textContent = user.login;
    document.getElementById('userSince').textContent = `Зарегистрирован: ${new Date(user.create_at).toLocaleDateString()}`;
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
                <button class="btn-primary" onclick="resetAndOpenTaskModal()">
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
        if (!task.due_date) return false;
        const today = new Date().toDateString();
        const taskDate = new Date(task.due_date).toDateString();
        return taskDate === today;
    }).length;

    document.getElementById('totalTasks').textContent = totalTasks;
    document.getElementById('activeTasks').textContent = activeTasks;
    document.getElementById('completedTasks').textContent = completedTasks;
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

// Функция обновления задачи
async function updateTask(taskId, taskData) {
    try {
        const response = await fetch(`${API_BASE}/tasks/${taskId}`, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: JSON.stringify(taskData)
        });

        if (response.ok) {
            showNotification('Задача успешно обновлена', 'success');
            closeTaskModal();
            await loadTasks();
        } else {
            throw new Error('Failed to update task');
        }
    } catch (error) {
        console.error('Failed to update task:', error);
        showNotification('Ошибка обновления задачи', 'error');
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

// Функция редактирования задачи
async function editTask(taskId) {
    try {
        // Загружаем полные данные задачи с сервера
        const response = await fetch(`${API_BASE}/tasks/${taskId}`, {
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to load task data');
        }

        const task = await response.json();
        currentEditingTaskId = taskId;

        // Заполняем форму данными задачи
        document.getElementById('taskTitle').value = task.title;
        document.getElementById('taskDescription').value = task.description || '';
        document.getElementById('taskPriority').value = task.priority;

        if (task.due_date) {
            const date = new Date(task.due_date);
            const formattedDate = date.toISOString().split('T')[0];
            document.getElementById('taskDueDate').value = formattedDate;
        } else {
            document.getElementById('taskDueDate').value = '';
        }

        // Меняем заголовок и текст кнопки
        document.querySelector('#taskModal .modal-header h3').textContent = 'Редактировать задачу';
        document.querySelector('#taskModal button[type="submit"]').textContent = 'Сохранить изменения';

        // Открываем модальное окно
        openTaskModal();
    } catch (error) {
        console.error('Failed to load task for editing:', error);
        showNotification('Ошибка загрузки задачи', 'error');
    }
}

// Настройка обработчиков событий
function setupEventListeners() {
    // Форма задачи
    document.getElementById('taskForm').addEventListener('submit', async (e) => {
        e.preventDefault();

        const taskData = {
            title: document.getElementById('taskTitle').value,
            description: document.getElementById('taskDescription').value,
            priority: document.getElementById('taskPriority').value,
            due_date: document.getElementById('taskDueDate').value || null
        };

        if (currentEditingTaskId) {
            await updateTask(currentEditingTaskId, taskData);
        } else {
            await createTask(taskData);
        }
    });

    // Фильтрация задач
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            filterTasks(btn.dataset.filter);
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
        task.style.display = (title.includes(searchTerm) || description.includes(searchTerm)) ? 'block' : 'none';
    });
}

// Навигация по секциям
function showSection(sectionName) {
    document.querySelectorAll('.content-section').forEach(section => {
        section.classList.remove('active');
    });
    document.getElementById(`${sectionName}-section`).classList.add('active');

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
    resetTaskModal();
}

function resetTaskModal() {
    document.getElementById('taskForm').reset();
    currentEditingTaskId = null;
    document.querySelector('#taskModal .modal-header h3').textContent = 'Новая задача';
    document.querySelector('#taskModal button[type="submit"]').textContent = 'Создать задачу';
}

function resetAndOpenTaskModal() {
    resetTaskModal();
    openTaskModal();
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
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;

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
    setTimeout(() => notification.remove(), 5000);
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