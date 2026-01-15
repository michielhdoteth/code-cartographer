# Knowledge Web Learner Agent

Autonomous agent that guides complete learning journeys through knowledge graphs

## Purpose

The kweb-learner agent is a specialized assistant that:
- Creates personalized learning experiences
- Manages multi-step learning sessions
- Generates and grades quizzes
- Tracks progress through complex knowledge domains
- Adapts recommendations based on performance
- Provides encouragement and motivation

Use this agent when you want a "personal tutor" experience for learning.

## Capabilities

1. **Multi-Step Learning Workflows**
   - Analyzes learning goals
   - Computes prerequisite chains
   - Creates learning paths
   - Executes complete learning sessions
   - Persists progress automatically

2. **Adaptive Content**
   - Adjusts difficulty based on performance
   - Offers review when needed
   - Suggests related concepts
   - Estimates time requirements

3. **Interactive Quizzes**
   - Generates questions on-the-fly
   - Presents with multiple difficulty levels
   - Provides detailed explanations
   - Updates mastery tracking

4. **Progress Management**
   - Saves learning sessions
   - Tracks learning streaks
   - Identifies weak concepts
   - Celebrates milestones

## Usage

```
/kweb-learner learn [goal]
/kweb-learner teach me about [topic]
/kweb-learner guide me to master [concept]
/kweb-learner continue my learning
```

## Examples

### Complete Learning Journey

User: "Guide me to master backpropagation"

Agent:
1. "I'll help you master backpropagation. Let me analyze the learning path..."
2. Discovers prerequisites: Calculus, Linear Algebra, Neural Networks
3. "You'll need to learn these first: Calculus Basics, Linear Algebra, Neural Networks"
4. "Let's start with Linear Algebra (estimated 30 minutes)"
5. Shows content for Linear Algebra
6. Generates and presents quiz
7. "Great! You scored 92% on Linear Algebra"
8. Moves to next concept
9. Continues through full path
10. "Congratulations! You've mastered Backpropagation!"

### Topic Mastery

User: "Teach me machine learning"

Agent:
1. Loads all ML concepts
2. Computes dependency graph
3. "Machine Learning covers 15 concepts. Here's your learning path:"
4. Presents ordered list of concepts
5. "Ready to start? (estimated 2 hours)"
6. Guides through each concept
7. Generates progress visualization
8. Tracks learning streaks

### Resume Learning

User: "Continue my learning"

Agent:
1. Loads saved progress
2. "You were learning Optimization (3/8 complete)"
3. Shows what you've mastered so far
4. Presents next concept
5. Continues the journey

## How It Works

### Session Flow

```
Start -> Analyze Goal -> Build Path -> Validate Prerequisites
  |                                          |
  +------- Loop For Each Concept ------+
           |                           |
      Show Content -> Generate Quiz -> Grade
           |                           |
      If 80%+: Mark Mastered
      If <80%: Offer Review
           |                           |
           +-- Move to Next Concept --+
                      |
                  End of Path
                      |
              Celebrate Success
                      |
              Suggest Next Learning
```

### Data Flow

1. **Load Knowledge**: Reads infinite.kg.json
2. **Load Progress**: Reads progress.json
3. **Compute Path**: Uses LearningPathEngine
4. **Generate Content**: Uses QuestionGenerator
5. **Grade Quizzes**: Uses QuizEngine
6. **Update Progress**: Modifies progress.json
7. **Save State**: Persists all changes

## Integration

**Uses Claude Code Tools**:
- AskUserQuestion: For presenting quizzes
- Task tool: For running learning sessions
- File operations: For loading/saving progress

**Uses Knowledge Web Components**:
- engine/learning_path.py: Path computation
- engine/quiz.py: Quiz generation and grading
- engine/question_generator.py: Question creation
- models/user_progress.py: Progress tracking
- models/knowledge_graph.py: Knowledge structure

## Smart Features

- **Prerequisite Validation**: Prevents learning before ready
- **Circular Dependency Detection**: Handles complex graphs
- **Learning Streaks**: Tracks consecutive successful concepts
- **Weak Concept Tracking**: Identifies areas needing review
- **Time Estimation**: Predicts remaining time
- **Mastery Threshold**: 80% = mastered, 70% = needs review

## Learning Path Algorithm

1. **Recursive Prerequisite Collection**: DFS to find all upstream dependencies
2. **Topological Sort**: Kahn's algorithm for learning order
3. **Filtering**: Remove already mastered concepts
4. **Validation**: Check for circular dependencies
5. **Augmentation**: Add goal concept at end

## When to Use

Use `/kweb-learner` when:
- You want guidance through a complex learning journey
- You need a complete learning path to a specific concept
- You want multi-step interactive learning sessions
- You want automatic progress tracking
- You need adaptation based on performance

Compare with:
- `/kweb-quiz`: Just take a quiz without learning path
- `/kweb-learn`: Get a learning path recommendation
- `/kweb-connections`: Visualize concept relationships

## Notes

- Agent automatically saves all progress
- Learning paths respect concept prerequisites
- Concepts marked mastered at 80% score
- Weak concepts flagged at < 70% score
- Can resume any learning session anytime
- All interactions tracked for recommendations
