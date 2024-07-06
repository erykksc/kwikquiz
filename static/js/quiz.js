document.addEventListener('DOMContentLoaded', function() {
    const addQuestionBtn = document.getElementById('add-question-btn');
    const questionList = document.getElementById('questions-list');

    addQuestionBtn.addEventListener('click', addNewQuestion);
    questionList.addEventListener('click', handleQuestionListEvents);

    function addNewQuestion() {
        const questionNumber = document.querySelectorAll('.question-item').length + 1;
        const questionHTML = createQuestionHTML(questionNumber);
        questionList.insertAdjacentHTML('beforeend', questionHTML);
        updateDeleteButtons();
    }

    function createQuestionHTML(questionNumber) {
        return `
        <div class="question-item mb-4 p-4 border border-baby-pink rounded-lg">
          <label class="block text-gray-700 font-semibold mb-2">Question ${questionNumber}</label>
          <input type="text" name="question-${questionNumber}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Enter question text" required>
          <label class="block text-gray-700 font-semibold mb-2">Answer Options</label>
          <div class="answers-container answers-container-${questionNumber}">
            <!-- Answer Options will be added dynamically here -->
          </div>
          <button type="button" class="add-answer-btn px-4 py-2 bg-dark-green text-white rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green mb-2">Add Answer</button>
          <button type="button" class="delete-question-btn px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete Question</button>
        </div>
        `;
    }

    function handleQuestionListEvents(event) {
        if (event.target.classList.contains('add-answer-btn')) {
            addNewAnswer(event.target);
        } else if (event.target.classList.contains('add-answer-type-btn')) {
            handleAnswerTypeChange(event.target);
        } else if (event.target.classList.contains('delete-answer-btn')) {
            event.target.closest('.answer-option').remove();
        } else if (event.target.classList.contains('correct-answer-btn')) {
            toggleCorrectAnswer(event.target);
        } else if (event.target.classList.contains('delete-question-btn')) {
            deleteQuestion(event.target);
        }
    }

    function addNewAnswer(addAnswerBtn) {
        const questionItem = addAnswerBtn.closest('.question-item');
        const questionNumber = questionItem.querySelector('label').textContent.trim().replace('Question ', '');
        const answersContainer = questionItem.querySelector(`.answers-container-${questionNumber}`);
        const answerCount = answersContainer.querySelectorAll('.answer-option').length + 1;
        const newAnswerHTML = createAnswerHTML(questionNumber, answerCount);
        answersContainer.insertAdjacentHTML('beforeend', newAnswerHTML);
    }

    function createAnswerHTML(questionNumber, answerCount) {
        return `
          <div class="answer-option mb-2">
            <input type="text" name="answer-${questionNumber}-${answerCount}" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" placeholder="Option ${answerCount}" required>
            <div class="flex justify-between items-center mt-2">
                <div>
                    <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerCount}" data-type="text">Text</button>
                    <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerCount}" data-type="image">Image</button>
                    <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerCount}" data-type="latex">LaTeX</button>
                </div>
                <div>
                    <button type="button" class="correct-answer-btn text-sm px-2 py-1 bg-red-500 text-white rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-green-500">Incorrect</button>
                    <button type="button" class="delete-answer-btn text-sm px-2 py-1 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete</button>
                </div>
            </div>
            <input type="hidden" name="correct-answer-${questionNumber}-${answerCount}" value="Incorrect">
          </div>
        `;
    }

    function handleAnswerTypeChange(button) {
        const answerInputName = button.getAttribute('data-answer');
        const answerType = button.getAttribute('data-type');
        const answerInput = document.querySelector(`input[name="${answerInputName}"], textarea[name="${answerInputName}"]`);

        if (!answerInput) return;

        switch (answerType) {
            case 'text':
                answerInput.outerHTML = `<input type="text" name="${answerInputName}" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" placeholder="Enter text" required>`;
                break;
            case 'image':
                answerInput.outerHTML = `<input type="file" name="${answerInputName}" accept="image/*" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" required>`;
                break;
            case 'latex':
                answerInput.outerHTML = `<textarea name="${answerInputName}" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" placeholder="Enter LaTeX" required></textarea>`;
                break;
        }
    }

    function toggleCorrectAnswer(button) {
        const isCurrentlyCorrect = button.textContent === 'Correct';
        const newCorrectState = !isCurrentlyCorrect;

        button.classList.toggle('bg-red-500', isCurrentlyCorrect);
        button.classList.toggle('bg-dark-green', newCorrectState);
        button.textContent = newCorrectState ? 'Correct' : 'Incorrect';

        const hiddenInput = button.closest('.answer-option').querySelector('input[type="hidden"]');
        if (hiddenInput) {
            hiddenInput.value = newCorrectState ? 'Correct' : 'Incorrect';
        }
    }

    function deleteQuestion(button) {
        button.closest('.question-item').remove();
        updateQuestionNumbers();
        toggleDeleteButtons();
    }

    function updateDeleteButtons() {
        const deleteButtons = document.querySelectorAll('.delete-question-btn');
        deleteButtons.forEach(button => {
            button.addEventListener('click', () => deleteQuestion(button));
        });
        toggleDeleteButtons();
    }

    function toggleDeleteButtons() {
        const deleteButtons = document.querySelectorAll('.delete-question-btn');
        const shouldHide = deleteButtons.length <= 1;
        deleteButtons.forEach(button => {
            button.classList.toggle('hidden', shouldHide);
        });
    }

    function updateQuestionNumbers() {
        const questionItems = document.querySelectorAll('.question-item');
        questionItems.forEach((item, index) => {
            const questionNumber = index + 1;
            updateQuestionLabelsAndInputs(item, questionNumber);
            updateAnswersContainer(item, questionNumber);
        });
    }

    function updateQuestionLabelsAndInputs(item, questionNumber) {
        const questionLabel = item.querySelector('label');
        questionLabel.textContent = `Question ${questionNumber}`;
        const inputs = item.querySelectorAll('input, textarea');
        inputs.forEach(input => {
            input.name = input.name.replace(/-\d+/, `-${questionNumber}`);
            input.id = input.id ? input.id.replace(/-\d+/, `-${questionNumber}`) : '';
            const label = item.querySelector(`label[for="${input.id}"]`);
            if (label) {
                label.setAttribute('for', input.id);
            }
        });
    }

    function updateAnswersContainer(item, questionNumber) {
        const answersContainer = item.querySelector('.answers-container');
        if (answersContainer) {
            answersContainer.className = `answers-container answers-container-${questionNumber}`;
        }
    }
});
