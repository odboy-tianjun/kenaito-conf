namespace Kenaito.Conf.Client;

public delegate void ChangeHandler(Dictionary<string, string> changed);

internal class Listener
{
    private readonly List<ChangeHandler> _handlers = new();

    public void OnChange(ChangeHandler handler)
    {
        _handlers.Add(handler);
    }

    public void Notify(Dictionary<string, string> changed)
    {
        foreach (var handler in _handlers)
        {
            try
            {
                handler(changed);
            }
            catch
            {
                // ignore handler errors
            }
        }
    }
}
