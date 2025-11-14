import json
import sys
from io import StringIO
import traceback


def handler(event, context):
    """
    Execute Python code in a sandboxed Lambda environment

    Event structure:
    {
        "code": "python code to execute",
        "input": {...},  # Input data available as 'input' variable
        "timeout": 60  # Optional timeout in seconds
    }
    """
    code = event.get('code', '')
    input_data = event.get('input', {})

    if not code:
        return {
            'statusCode': 400,
            'body': json.dumps({
                'success': False,
                'error': 'No code provided'
            })
        }

    try:
        # Capture stdout
        old_stdout = sys.stdout
        sys.stdout = captured_output = StringIO()

        # Create execution namespace
        namespace = {
            'input': input_data,
            '__builtins__': __builtins__,
            'json': json,
        }

        # Execute code
        exec(code, namespace)

        # Restore stdout
        sys.stdout = old_stdout
        output = captured_output.getvalue()

        # Extract result (look for 'result' variable or use last expression)
        result = namespace.get('result', None)

        return {
            'statusCode': 200,
            'body': json.dumps({
                'success': True,
                'result': result,
                'output': output,
                'executedAt': context.request_id
            })
        }

    except Exception as e:
        sys.stdout = old_stdout

        return {
            'statusCode': 500,
            'body': json.dumps({
                'success': False,
                'error': str(e),
                'traceback': traceback.format_exc()
            })
        }
