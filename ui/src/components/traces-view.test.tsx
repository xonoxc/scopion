import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TracesView } from '../components/traces-view'

interface TraceInfo {
  id: string
  name: string
  service: string
  duration: number
  spans: number
  timestamp: string
  has_error: boolean
}

// Mock the useTraces hook
vi.mock('../hooks/use-traces', () => ({
  useTraces: vi.fn(),
}))

// Mock the useNavigate hook
vi.mock('@tanstack/react-router', () => ({
  useNavigate: vi.fn(),
}))

// Mock the cn utility
vi.mock('../lib/utils', () => ({
  cn: (...classes: (string | undefined | null | boolean)[]) => classes.filter(Boolean).join(' '),
}))

import { useTraces } from '../hooks/use-traces'
import { useNavigate } from '@tanstack/react-router'

const mockUseTraces = vi.mocked(useTraces)
const mockUseNavigate = vi.mocked(useNavigate)

describe('TracesView', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    })
    mockUseNavigate.mockReturnValue(vi.fn())
  })

  const renderComponent = (serviceFilter: string | null = null) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <TracesView serviceFilter={serviceFilter} />
      </QueryClientProvider>
    )
  }

  it('shows loading state when data is loading', () => {
    mockUseTraces.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
      isError: false,
      refetch: vi.fn(),
    } as any)

    renderComponent()

    expect(screen.getByText('Loading traces...')).toBeInTheDocument()
  })

  it('renders traces when data is available', () => {
    const mockTraces: TraceInfo[] = [
      {
        id: 'trace-1',
        name: 'HTTP Request',
        service: 'web-service',
        duration: 150,
        spans: 3,
        timestamp: '2023-01-01T10:00:00Z',
        has_error: false,
      },
      {
        id: 'trace-2',
        name: 'Database Query',
        service: 'db-service',
        duration: 200,
        spans: 2,
        timestamp: '2023-01-01T09:59:00Z',
        has_error: true,
      },
    ]

    mockUseTraces.mockReturnValue({
      data: mockTraces,
      isLoading: false,
      error: null,
      isError: false,
      refetch: vi.fn(),
    } as any)

    renderComponent()

    expect(screen.getByText('2 traces')).toBeInTheDocument()
    expect(screen.getByText('HTTP Request')).toBeInTheDocument()
    expect(screen.getByText('Database Query')).toBeInTheDocument()
    expect(screen.getByText('web-service')).toBeInTheDocument()
    expect(screen.getByText('db-service')).toBeInTheDocument()
  })

  it('filters traces by service when serviceFilter is provided', () => {
    const mockTraces: TraceInfo[] = [
      {
        id: 'trace-1',
        name: 'HTTP Request',
        service: 'web-service',
        duration: 150,
        spans: 3,
        timestamp: '2023-01-01T10:00:00Z',
        has_error: false,
      },
      {
        id: 'trace-2',
        name: 'Database Query',
        service: 'db-service',
        duration: 200,
        spans: 2,
        timestamp: '2023-01-01T09:59:00Z',
        has_error: true,
      },
    ]

    mockUseTraces.mockReturnValue({
      data: mockTraces,
      isLoading: false,
      error: null,
      isError: false,
      refetch: vi.fn(),
    } as any)

    renderComponent('web-service')

    expect(screen.getByText('1 traces • filtered by web-service')).toBeInTheDocument()
    expect(screen.getByText('HTTP Request')).toBeInTheDocument()
    expect(screen.queryByText('Database Query')).not.toBeInTheDocument()
  })

  it('shows no traces when filtered result is empty', () => {
    const mockTraces: TraceInfo[] = [
      {
        id: 'trace-1',
        name: 'HTTP Request',
        service: 'web-service',
        duration: 150,
        spans: 3,
        timestamp: '2023-01-01T10:00:00Z',
        has_error: false,
      },
    ]

    mockUseTraces.mockReturnValue({
      data: mockTraces,
      isLoading: false,
      error: null,
      isError: false,
      refetch: vi.fn(),
    } as any)

    renderComponent('non-existent-service')

    expect(screen.getByText('0 traces • filtered by non-existent-service')).toBeInTheDocument()
  })

  it('displays error indicator for traces with errors', () => {
    const mockTraces: TraceInfo[] = [
      {
        id: 'trace-1',
        name: 'HTTP Request',
        service: 'web-service',
        duration: 150,
        spans: 3,
        timestamp: '2023-01-01T10:00:00Z',
        has_error: true,
      },
    ]

    mockUseTraces.mockReturnValue({
      data: mockTraces,
      isLoading: false,
      error: null,
      isError: false,
      refetch: vi.fn(),
    } as any)

    renderComponent()

    expect(screen.getByText('ERROR')).toBeInTheDocument()
  })

  it('navigates to trace detail on click', () => {
    const mockNavigate = vi.fn()
    mockUseNavigate.mockReturnValue(mockNavigate)

    const mockTraces: TraceInfo[] = [
      {
        id: 'trace-1',
        name: 'HTTP Request',
        service: 'web-service',
        duration: 150,
        spans: 3,
        timestamp: '2023-01-01T10:00:00Z',
        has_error: false,
      },
    ]

    mockUseTraces.mockReturnValue({
      data: mockTraces,
      isLoading: false,
      error: null,
      isError: false,
      refetch: vi.fn(),
    } as any)

    renderComponent()

    const traceElement = screen.getByText('HTTP Request').closest('div')
    traceElement?.click()

    expect(mockNavigate).toHaveBeenCalledWith({ to: '/traces/trace-1' })
  })
})